## NAT library for Golang

This library is for NAT port mapping *only*!
* Currently Support UPnP for IGD v1 & v2
* Support NAT-PMP (by plan)
* Support DMZ (by plan)

### Installation

```
Run go get https://github.com/symphonyprotocol/nat
```

### Usage
1. UPnP port mapping

```
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
		ip, _ := upnp.GetExternalIPAddress(client)
		fmt.Println(ip)
		if ok := upnp.AddPortMapping(ips[0], 1234, 1234, "UDP", client); ok {
			fmt.Println("add UDP success")
		}
		if ok := upnp.AddPortMapping(ips[0], 1234, 1234, "TCP", client); ok {
			fmt.Println("add UDP success")
		}
		if ok := upnp.DeletePortMapping(1234, "UDP", client); ok {
			fmt.Println("delete success")
		}
	}

}
```
### module support 

* In progress
