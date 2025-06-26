#!/bin/bash
socat -d -d -d TCP4-LISTEN:27015,reuseaddr,fork UNIX-CONNECT:${USBMUXD_SOCKET_ADDRESS}&

usbmuxd
#sleep 5
##idevicepair pair
#ideviceinfo
##idevicepair pair
