package main

import (
	"fmt"

	"github.com/f5devcentral/go-bigip"
)

func main() {

	fmt.Println("Attempting to connect...")
	// Establish our session to the BIG-IP
	f5 := bigip.NewSession("10.192.74.73", "admin", "admin", nil)

	// Iterate over all the virtual servers, and display their names.
	vservers, err := f5.VirtualServers()
	if err != nil {
		panic(err.Error())
	}

	for _, vs := range vservers.VirtualServers {
		fmt.Printf("Name: %s\n, %s\n", vs.Name, vs.Pool)
		//vs.Description = "Modified Sanjay Shitole"
		f5.ModifyVirtualServer(vs.Name, &vs)

	}

	vaddrs, err := f5.VirtualAddresses()
	if err != nil {
		panic(err.Error())
	}
	for _, va := range vaddrs.VirtualAddresses {
		fmt.Printf("VA: %+v\n", va)
	}
}
