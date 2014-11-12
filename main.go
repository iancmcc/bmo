package main

import (
	"fmt"

	"github.com/huin/goupnp"
	"github.com/huin/goupnp/soap"
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

func discover(target string, iface chan goupnp.MaybeRootDevice) {
	devices, _ := goupnp.DiscoverDevices(target)
	for _, device := range devices {
		iface <- device
	}
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
	err := client.SOAPClient.PerformAction(URN_BasicEvent_1, "GetBinaryState", request, response)
	fmt.Printf("THEERR: %+v\n", err)
	binaryState, err := soap.UnmarshalString(response.BinaryState)
	fmt.Printf("BS: %+v\n", response.BinaryState)
	fmt.Printf("Err: %+v\n", err)
	return binaryState
}

func main() {
	channel := make(chan goupnp.MaybeRootDevice)

	go discover(URN_Switch, channel)
	//go discover(URN_Motion, channel)
	//go discover(URN_LightSwitch, channel)
	//go discover(URN_Insight, channel)

	for device := range channel {
		sw := &WeMoDevice{device.Root}
		fmt.Println(sw.GetBasicEventClient().GetBinaryState())
		//dev := device.Root.Device
		//svc := dev.FindService(URN_BasicEvent_1)
		//fmt.Printf("%+v", svc)
		//client := goupnp.ServiceClient{
		//	SOAPClient: svc.NewSOAPClient(),
		//	RootDevice: device.Root,
		//	Service:    svc,
		//}
		//scpd, _ := svc.RequestSCDP()
		//for _, action := range scpd.Actions {
		//	fmt.Printf(" * %s\n", action.Name)
		//	for _, arg := range action.Arguments {
		//		var varDesc string
		//		if stateVar := scpd.GetStateVariable(arg.RelatedStateVariable); stateVar != nil {
		//			varDesc = fmt.Sprintf(" (%s)", stateVar.DataType.Name)
		//		}
		//		fmt.Printf("    * [%s] %s%s\n", arg.Direction, arg.Name, varDesc)
		//	}
		//}
	}
}
