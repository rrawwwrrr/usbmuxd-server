#!/bin/bash
export USBMUXD_SOCKET_ADDRESS=/var/run/usbmuxd
socat -d -d -d TCP4-LISTEN:27015,reuseaddr,fork UNIX-CONNECT:${USBMUXD_SOCKET_ADDRESS}&

usbmuxd
#./goios.sh
#sleep 5
##idevicepair pair
#ideviceinfo
##idevicepair pair
