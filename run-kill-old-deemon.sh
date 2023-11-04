ps ax | grep kubevpn | grep -v grep | grep -v dlv | awk '{ print "kill "$1}' | sudo sh ; ps ax | grep kubevpn
