package goios

import (
	"context"
	"goios-peer/tools"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/imagemounter"
	"github.com/danielpaulus/go-ios/ios/testmanagerd"
	"github.com/danielpaulus/go-ios/ios/tunnel"
	log "github.com/sirupsen/logrus"
)

var tm *tunnel.TunnelManager
var tunnelInfoPort = 60000

func Start() {
	pm, err := tunnel.NewPairRecordManager(".")
	tools.ExitIfError("could not creat pair record manager", err)
	tm = tunnel.NewTunnelManager(pm, false)
	go startTunnel(context.TODO())
	time.Sleep(4 * time.Second)
	devices, err := ios.ListDevices()
	if err != nil {
		log.Fatal(err)
	}
	for i := range devices.DeviceList {
		log.Print(devices.DeviceList[i].Properties.SerialNumber)
	}
	device := devices.DeviceList[0]
	err = imagemounter.MountImage(device, "")
	if err != nil {
		log.Fatal(err)
	}
	runWda(device)
	//err = instruments.StartMJPEGStreamingServer(device, "3333")
	//if err != nil {
	//	log.Fatal(err)
	//}

}

func runWda(device ios.DeviceEntry) {
	bundleID := "com.facebook.WebDriverAgentRunner.xctrunner"
	testbundleID := "com.facebook.WebDriverAgentRunner.xctrunner"
	xctestconfig := "WebDriverAgentRunner.xctest"
	wdaargs := []string{}

	var writer = io.Discard

	errorChannel := make(chan error)
	defer close(errorChannel)
	ctx, stopWda := context.WithCancel(context.Background())

	go func() {
		info, _ := tm.FindTunnel(device.Properties.SerialNumber)
		log.Printf(info.Udid)
		device.UserspaceTUNPort = info.UserspaceTUNPort
		device.UserspaceTUN = info.UserspaceTUN
		device = deviceWithRsdProvider(device, info.Address, info.RsdPort)
		_, err := testmanagerd.RunTestWithConfig(ctx, testmanagerd.TestConfig{
			BundleId:           bundleID,
			TestRunnerBundleId: testbundleID,
			XctestConfigName:   xctestconfig,
			Args:               wdaargs,
			Device:             device,
			Listener:           testmanagerd.NewTestListener(writer, writer, os.TempDir()),
		})
		if err != nil {
			errorChannel <- err
		}
		stopWda()
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-errorChannel:
		log.Println("Failed running WDA")
		stopWda()
		os.Exit(1)
	case <-ctx.Done():
		log.Println("WDA process ended unexpectedly")
		os.Exit(1)
	case signal := <-c:
		log.Printf("os signal:%d received, closing...", signal)
		stopWda()
	}
	log.Print("Done Closing")
}

func startTunnel(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := tm.UpdateTunnels(ctx)
				if err != nil {
					log.WithError(err).Warn("failed to update tunnels")
				}
			}
		}
	}()

	go func() {
		err := tunnel.ServeTunnelInfo(tm, tunnelInfoPort)
		if err != nil {
			tools.ExitIfError("failed to start tunnel server", err)
		}
	}()
	log.Info("Tunnel server started")
	<-ctx.Done()
}

func deviceWithRsdProvider(device ios.DeviceEntry, address string, rsdPort int) ios.DeviceEntry {
	udid := device.Properties.SerialNumber
	rsdService, err := ios.NewWithAddrPortDevice(address, rsdPort, device)
	tools.ExitIfError("could not connect to RSD", err)
	defer rsdService.Close()
	rsdProvider, err := rsdService.Handshake()
	device1, err := ios.GetDeviceWithAddress(udid, address, rsdProvider)
	device1.UserspaceTUN = device.UserspaceTUN
	device1.UserspaceTUNPort = device.UserspaceTUNPort
	tools.ExitIfError("error getting devicelist", err)

	return device1
}
