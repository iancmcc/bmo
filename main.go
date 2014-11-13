package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/iancmcc/goupnp"
	"github.com/iancmcc/goupnp/soap"
	"github.com/jessevdk/go-flags"
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

type Switch interface {
	On()
	Off()
	Toggle() int
}

type BasicEventClient struct {
	goupnp.ServiceClient
}

func (d *WeMoDevice) GetBasicEventClient() *BasicEventClient {
	svc := d.Root.Device.FindService(URN_BasicEvent_1)[0]
	return &BasicEventClient{goupnp.ServiceClient{
		SOAPClient: svc.NewSOAPClient(),
		RootDevice: d.Root,
		Service:    svc,
	}}
}

func (client *BasicEventClient) GetBinaryState() int {
	request := interface{}(nil)
	response := &struct {
		BinaryState string
	}{}
	if err := client.SOAPClient.PerformAction(URN_BasicEvent_1, "GetBinaryState", request, response); err != nil {
		log.Printf("Error: %+v\n", err)
		return -1
	}
	binaryState, err := soap.UnmarshalString(response.BinaryState)
	if err != nil {
		log.Printf("Error: %+v\n", err)
		return -1
	}
	result, err := strconv.ParseInt(binaryState, 10, 0)
	if err != nil {
		log.Printf("Error: %+v\n", err)
		return -1
	}
	return int(result)
}

func (client *BasicEventClient) SetBinaryState(state int) {
	request := &struct {
		BinaryState string
	}{BinaryState: fmt.Sprintf("%d", state)}
	response := interface{}(nil)
	if err := client.SOAPClient.PerformAction(URN_BasicEvent_1, "SetBinaryState", request, response); err != nil {
		log.Printf("Error: %+v\n", err)
		return
	}
}

func (d *WeMoDevice) On() {
	d.GetBasicEventClient().SetBinaryState(1)
}

func (d *WeMoDevice) Off() {
	d.GetBasicEventClient().SetBinaryState(0)
}

func (d *WeMoDevice) Toggle() {
	client := d.GetBasicEventClient()
	client.SetBinaryState(1 - client.GetBinaryState())
}

type Options struct {
	// Example of verbosity with level
	Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
}

var (
	options Options
	parser  = flags.NewParser(&options, flags.Default)
)

func main() {
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	} else {
		fmt.Printf("%+v", options)
	}

	switches := discover(URN_Switch)
	motions := discover(URN_Motion)
	lightswitches := discover(URN_LightSwitch)
	insights := discover(URN_Insight)

	for device := range mergeDevices(switches, motions, lightswitches, insights) {
		sw := &WeMoDevice{device.Root}
		if sw.Root.Device.FriendlyName == "Bedroom Lights" {
			sw.GetBasicEventClient().SetBinaryState(1)
			return
		}
	}

}
