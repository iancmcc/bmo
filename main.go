package main

import (
	"fmt"
	"log"

	"github.com/iancmcc/goupnp"
	"github.com/iancmcc/goupnp/soap"
)

const (
	URN_Switch      = "urn:Belkin:device:controllee:1"
	URN_Motion      = "urn:Belkin:device:sensor:1"
	URN_LightSwitch = "urn:Belkin:device:lightswitch:1"
	URN_Insight     = "urn:Belkin:device:insight:1"
)

const (
	URN_BasicEvent_1 = "urn:Belkin:service:basicevent:1"
)

func discover(target string) <-chan goupnp.MaybeRootDevice {
	devices, err := goupnp.DiscoverDevices(target)
	if err != nil {
		log.Printf("Unable to discover %s devices: %v\n", target, err)
		out := make(chan goupnp.MaybeRootDevice)
		close(out)
		return out
	}
	return devices
}

type WeMoDevice struct {
	Root *goupnp.RootDevice
}

func (d *WeMoDevice) GetBasicEventClient() *BasicEventClient {
	svc := d.Root.Device.FindService(URN_BasicEvent_1)[0]
	return &BasicEventClient{goupnp.ServiceClient{
		SOAPClient: svc.NewSOAPClient(),
		RootDevice: d.Root,
		Service:    svc,
	}}
}

type BasicEventClient struct {
	goupnp.ServiceClient
}

func (client *BasicEventClient) GetBinaryState() string {
	request := interface{}(nil)
	response := &struct {
		BinaryState string
	}{}
	if err := client.SOAPClient.PerformAction(URN_BasicEvent_1, "GetBinaryState", request, response); err != nil {
		fmt.Printf("THEERR: %+v\n", err)
		return ""
	}
	binaryState, err := soap.UnmarshalString(response.BinaryState)
	if err != nil {
		fmt.Printf("Err: %+v\n", err)
		return ""
	}
	fmt.Printf("BS: %+v\n", response.BinaryState)
	return binaryState
}

func main() {
	switches := discover(URN_Switch)
	motions := discover(URN_Motion)
	lightswitches := discover(URN_LightSwitch)
	insights := discover(URN_Insight)

	for device := range mergeDevices(switches, motions, lightswitches, insights) {
		fmt.Println(device)
		sw := &WeMoDevice{device.Root}
		fmt.Println(sw.Root.Device.FriendlyName)
		fmt.Println(sw.GetBasicEventClient().GetBinaryState())
	}
}
