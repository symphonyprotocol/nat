package main

import (
	"fmt"
	"net"

	"github.com/symphonyprotocol/nat"
)

func main() {
	ips, _ := nat.IntranetIP()
	fmt.Println(ips)
	ip, _ := nat.GetOutbountIP()
	ipAddr := net.ParseIP(ip)
	fmt.Println(ipAddr.Mask(ipAddr.DefaultMask()))
	// client, _ := upnp.NewUPnPClient()
	// if ok := client.Discover(); ok {
	// 	index := 0
	// 	for {
	// 		fmt.Println("--------", index)
	// 		protocol, ip, port, err := upnp.GetGenericPortMappingEntry(index, client)
	// 		if port == 0 {
	// 			break
	// 		}
	// 		fmt.Println(protocol, ip, port, err)
	// 		index++
	// 	}
	// 	//ip, _ := upnp.GetExternalIPAddress(client)
	// 	//fmt.Println(ip)
	// 	//if ok := upnp.AddPortMapping(ips[0], 1234, 1234, "UDP", client); ok {
	// 	//	fmt.Println("add UDP success")
	// 	//}
	// 	//if ok := upnp.AddPortMapping(ips[0], 1234, 1234, "TCP", client); ok {
	// 	//		fmt.Println("add UDP success")
	// 	//}
	// 	//if ok := upnp.DeletePortMapping(1234, "UDP", client); ok {
	// 	//	fmt.Println("delete success")
	// 	//}
	// }

}
