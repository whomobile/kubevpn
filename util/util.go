package util

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	dockerterm "github.com/moby/term"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	json2 "k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	runtimeresource "k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"strconv"

	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/remotecommand"
	clientgowatch "k8s.io/client-go/tools/watch"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/client-go/util/retry"
	"k8s.io/kubectl/pkg/cmd/exec"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

func WaitResource(clientset *kubernetes.Clientset, getter cache.Getter, namespace, apiVersion, kind string, list metav1.ListOptions, checker func(interface{}) bool) error {
	groupResources, _ := restmapper.GetAPIGroupResources(clientset)
	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)
	groupVersionKind := schema.FromAPIVersionAndKind(apiVersion, kind)
	mapping, err := mapper.RESTMapping(groupVersionKind.GroupKind(), groupVersionKind.Version)
	if err != nil {
		log.Error(err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	watchlist := cache.NewFilteredListWatchFromClient(
		getter,
		mapping.Resource.Resource,
		namespace,
		func(options *metav1.ListOptions) {
			options.LabelSelector = list.LabelSelector
			options.FieldSelector = list.FieldSelector
			options.Watch = list.Watch
		},
	)

	preConditionFunc := func(store cache.Store) (bool, error) {
		if len(store.List()) == 0 {
			return false, nil
		}
		for _, p := range store.List() {
			if !checker(p) {
				return false, nil
			}
		}
		return true, nil
	}

	conditionFunc := func(e watch.Event) (bool, error) { return checker(e.Object), nil }

	object, err := scheme.Scheme.New(mapping.GroupVersionKind)
	if err != nil {
		return err
	}

	event, err := clientgowatch.UntilWithSync(ctx, watchlist, object, preConditionFunc, conditionFunc)
	if err != nil {
		log.Infof("wait to ready failed, error: %v, event: %v", err, event)
		return err
	}
	return nil
}

func GetAvailablePortOrDie() int {
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", "0.0.0.0"))
	if err != nil {
		log.Fatal(err)
	}
	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func GetAvailableUDPPortOrDie() int {
	address, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:0", "0.0.0.0"))
	if err != nil {
		log.Fatal(err)
	}
	listener, err := net.ListenUDP("udp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	return listener.LocalAddr().(*net.UDPAddr).Port
}

func WaitPod(clientset *kubernetes.Clientset, namespace string, list metav1.ListOptions, checker func(*v1.Pod) bool) error {
	return WaitResource(
		clientset,
		clientset.CoreV1().RESTClient(),
		namespace,
		"v1",
		"Pod",
		list,
		func(i interface{}) bool { return checker(i.(*v1.Pod)) },
	)
}

func PortForwardPod(config *rest.Config, clientset *rest.RESTClient, podName, namespace, portPair string, readyChan, stopChan chan struct{}) error {
	url := clientset.
		Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward").
		URL()
	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		log.Error(err)
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)
	p := []string{portPair}
	forwarder, err := NewOnAddresses(dialer, []string{"0.0.0.0"}, p, stopChan, readyChan, os.Stdout, os.Stderr)
	if err != nil {
		log.Error(err)
		return err
	}

	if err = forwarder.ForwardPorts(); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func GetTopController(factory cmdutil.Factory, clientset *kubernetes.Clientset, namespace, serviceName string) (controller ResourceTupleWithScale) {
	object, err := GetUnstructuredObject(factory, namespace, serviceName)
	if err != nil {
		return
	}
	asSelector, _ := metav1.LabelSelectorAsSelector(GetLabelSelector(object))
	podList, _ := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: asSelector.String(),
	})
	if len(podList.Items) == 0 {
		return
	}
	of := metav1.GetControllerOf(&podList.Items[0])
	for of != nil {
		object, err = GetUnstructuredObject(factory, namespace, fmt.Sprintf("%s/%s", of.Kind, of.Name))
		if err != nil {
			return
		}
		controller.Resource = strings.ToLower(of.Kind) + "s"
		controller.Name = of.Name
		controller.Scale = GetScale(object)
		of = GetOwnerReferences(object)
	}
	return
}

func UpdateReplicasScale(clientset *kubernetes.Clientset, namespace string, controller ResourceTupleWithScale) {
	err := retry.OnError(
		retry.DefaultRetry,
		func(err error) bool { return err != nil },
		func() error {
			result := &autoscalingv1.Scale{}
			err := clientset.AppsV1().RESTClient().Put().
				Namespace(namespace).
				Resource(controller.Resource).
				Name(controller.Name).
				SubResource("scale").
				VersionedParams(&metav1.UpdateOptions{}, scheme.ParameterCodec).
				Body(&autoscalingv1.Scale{
					ObjectMeta: metav1.ObjectMeta{
						Name:      controller.Name,
						Namespace: namespace,
					},
					Spec: autoscalingv1.ScaleSpec{
						Replicas: int32(controller.Scale),
					},
				}).
				Do(context.Background()).
				Into(result)
			return err
		})
	if err != nil {
		log.Errorf("update scale: %s-%s's replicas to %d failed, error: %v", controller.Resource, controller.Name, controller.Scale, err)
	}
}

func Shell(clientset *kubernetes.Clientset, restclient *rest.RESTClient, config *rest.Config, podName, namespace, cmd string) (string, error) {
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})

	if err != nil {
		return "", err
	}
	if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
		err = fmt.Errorf("cannot exec into a container in a completed pod; current phase is %s", pod.Status.Phase)
		return "", err
	}
	containerName := pod.Spec.Containers[0].Name
	stdin, _, stderr := dockerterm.StdStreams()

	stdoutBuf := bytes.NewBuffer(nil)
	stdout := io.MultiWriter(stdoutBuf)
	StreamOptions := exec.StreamOptions{
		Namespace:     namespace,
		PodName:       podName,
		ContainerName: containerName,
		IOStreams:     genericclioptions.IOStreams{In: stdin, Out: stdout, ErrOut: stderr},
	}
	Executor := &exec.DefaultRemoteExecutor{}
	// ensure we can recover the terminal while attached
	tt := StreamOptions.SetupTTY()

	var sizeQueue remotecommand.TerminalSizeQueue
	if tt.Raw {
		// this call spawns a goroutine to monitor/update the terminal size
		sizeQueue = tt.MonitorSize(tt.GetSize())

		// unset p.Err if it was previously set because both stdout and stderr go over p.Out when tty is
		// true
		StreamOptions.ErrOut = nil
	}

	fn := func() error {
		req := restclient.Post().
			Resource("pods").
			Name(pod.Name).
			Namespace(pod.Namespace).
			SubResource("exec")
		req.VersionedParams(&v1.PodExecOptions{
			Container: containerName,
			Command:   []string{"sh", "-c", cmd},
			Stdin:     StreamOptions.Stdin,
			Stdout:    StreamOptions.Out != nil,
			Stderr:    StreamOptions.ErrOut != nil,
			TTY:       tt.Raw,
		}, scheme.ParameterCodec)
		return Executor.Execute("POST", req.URL(), config, StreamOptions.In, StreamOptions.Out, StreamOptions.ErrOut, tt.Raw, sizeQueue)
	}

	err = tt.Safe(fn)
	return strings.TrimRight(stdoutBuf.String(), "\n"), err
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func GetUnstructuredObject(f cmdutil.Factory, namespace string, workloads string) (k8sruntime.Object, error) {
	do := f.NewBuilder().
		Unstructured().
		NamespaceParam(namespace).DefaultNamespace().AllNamespaces(false).
		ResourceTypeOrNameArgs(true, workloads).
		ContinueOnError().
		Latest().
		Flatten().
		TransformRequests(func(req *rest.Request) { req.Param("includeObject", "Object") }).
		Do()
	if err := do.Err(); err != nil {
		log.Warn(err)
		return nil, err
	}
	infos, err := do.Infos()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if len(infos) == 0 {
		return nil, errors.New("Not found")
	}
	return infos[0].Object, err
}

func GetLabelSelector(object k8sruntime.Object) *metav1.LabelSelector {
	l := &metav1.LabelSelector{}

	printer, _ := printers.NewJSONPathPrinter("{.spec.selector}")
	buf := bytes.NewBuffer([]byte{})
	if err := printer.PrintObj(object, buf); err != nil {
		pathPrinter, _ := printers.NewJSONPathPrinter("{.metadata.labels}")
		_ = pathPrinter.PrintObj(object, buf)
	}
	err := json2.Unmarshal([]byte(buf.String()), l)
	if err != nil || len(l.MatchLabels) == 0 {
		m := map[string]string{}
		_ = json2.Unmarshal([]byte(buf.String()), &m)
		if len(m) != 0 {
			l = &metav1.LabelSelector{MatchLabels: m}
		}
	}
	return l
}

func GetPorts(object k8sruntime.Object) []v1.ContainerPort {
	var result []v1.ContainerPort
	replicasetPortPrinter, _ := printers.NewJSONPathPrinter("{.spec.template.spec.containers[0].ports}")
	servicePortPrinter, _ := printers.NewJSONPathPrinter("{.spec.ports}")
	buf := bytes.NewBuffer([]byte{})
	err := replicasetPortPrinter.PrintObj(object, buf)
	if err != nil {
		_ = servicePortPrinter.PrintObj(object, buf)
		var ports []v1.ServicePort
		_ = json2.Unmarshal([]byte(buf.String()), &ports)
		for _, port := range ports {
			val := port.TargetPort.IntVal
			if val == 0 {
				val = port.Port
			}
			result = append(result, v1.ContainerPort{
				Name:          port.Name,
				ContainerPort: val,
				Protocol:      port.Protocol,
			})
		}
	} else {
		_ = json2.Unmarshal([]byte(buf.String()), &result)
	}
	return result
}

func GetOwnerReferences(object k8sruntime.Object) *metav1.OwnerReference {
	printer, _ := printers.NewJSONPathPrinter("{.metadata.ownerReferences}")
	buf := bytes.NewBuffer([]byte{})
	if err := printer.PrintObj(object, buf); err != nil {
		return nil
	}
	var refs []metav1.OwnerReference
	_ = json2.Unmarshal([]byte(buf.String()), &refs)
	for i := range refs {
		if refs[i].Controller != nil && *refs[i].Controller {
			return &refs[i]
		}
	}
	return nil
}

func GetScale(object k8sruntime.Object) int {
	printer, _ := printers.NewJSONPathPrinter("{.spec.replicas}")
	buf := bytes.NewBuffer([]byte{})
	if err := printer.PrintObj(object, buf); err != nil {
		return 0
	}
	if atoi, err := strconv.Atoi(buf.String()); err == nil {
		return atoi
	}
	return 0
}

func DeletePod(clientset *kubernetes.Clientset, namespace, podName string) {
	zero := int64(0)
	err := clientset.CoreV1().Pods(namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{
		GracePeriodSeconds: &zero,
	})
	if err != nil && k8serrors.IsNotFound(err) {
		log.Infof("not found shadow pod: %s, no need to delete it", podName)
	}
}

// TopLevelControllerSet record every pod's top level controller, like pod controllerBy replicaset, replicaset controllerBy deployment
var TopLevelControllerSet []ResourceTupleWithScale

type ResourceTuple struct {
	Resource string
	Name     string
}

type ResourceTupleWithScale struct {
	Resource string
	Name     string
	Scale    int
}

// splitResourceTypeName handles type/name resource formats and returns a resource tuple
// (empty or not), whether it successfully found one, and an error
func SplitResourceTypeName(s string) (ResourceTuple, bool, error) {
	if !strings.Contains(s, "/") {
		return ResourceTuple{}, false, nil
	}
	seg := strings.Split(s, "/")
	if len(seg) != 2 {
		return ResourceTuple{}, false, fmt.Errorf("arguments in resource/name form may not have more than one slash")
	}
	resource, name := seg[0], seg[1]
	if len(resource) == 0 || len(name) == 0 || len(runtimeresource.SplitResourceArgument(resource)) != 1 {
		return ResourceTuple{}, false, fmt.Errorf("arguments in resource/name form must have a single resource and name")
	}
	return ResourceTuple{Resource: resource, Name: name}, true, nil
}

func DeleteConfigMap(clientset *kubernetes.Clientset, namespace, configMapName string) {
	_ = clientset.CoreV1().ConfigMaps(namespace).Delete(context.Background(), configMapName, metav1.DeleteOptions{})
}

func BytesToInt(b []byte) uint32 {
	buffer := bytes.NewBuffer(b)
	var u uint32
	if err := binary.Read(buffer, binary.BigEndian, &u); err != nil {
		log.Warn(err)
	}
	return u
}
