package main

import (
	"fmt"
	"github.com/symphonyprotocol/nat"
	"github.com/symphonyprotocol/nat/upnp"
)

func main() {
	ips, _ := nat.IntranetIP()
	fmt.Println(ips)
	client, _ := upnp.NewUPnPClient()
	if ok := client.Discover(); ok {
		index := 0
		for {
			fmt.Println("--------", index)
			protocol, ip, port, err := upnp.GetGenericPortMappingEntry(index, client)
			if port == 0 {
				break
			}
			fmt.Println(protocol, ip, port, err)
			index++
		}
		//ip, _ := upnp.GetExternalIPAddress(client)
		//fmt.Println(ip)
		//if ok := upnp.AddPortMapping(ips[0], 1234, 1234, "UDP", client); ok {
		//	fmt.Println("add UDP success")
		//}
		//if ok := upnp.AddPortMapping(ips[0], 1234, 1234, "TCP", client); ok {
		//		fmt.Println("add UDP success")
		//}
		//if ok := upnp.DeletePortMapping(1234, "UDP", client); ok {
		//	fmt.Println("delete success")
		//}
	}

}
