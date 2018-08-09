package upnp

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	MAX_WAIT_TIME       = 5
	MAX_RETRY           = 3
	METHOD_SEARCH       = "M-SEARCH"
	BOARDCAST_ADDR      = "255.255.255.255:1900"
	SSDP_BOARDCAST_ADDR = "239.255.255.250:1900"
	SSDP_DISCOVER       = `"ssdp:discover"`
	SEARCH_TARGET       = "upnp:rootdevice"
	USN_ISG             = "uuid:upnp-InternetGatewayDevice"
	DEVICE_TYPE_ISG     = "urn:schemas-upnp-org:device:InternetGatewayDevice"
	DEVICE_TYPE_WD      = "urn:schemas-upnp-org:device:WANDevice"
	DEVICE_TYPE_WCD     = "urn:schemas-upnp-org:device:WANConnectionDevice"
	ST_WIPC             = "urn:schemas-upnp-org:service:WANIPConnection"
	ST_WPPPC            = "urn:schemas-upnp-org:service:WANPPPConnection"
)

type UPnPClient struct {
	connLock    sync.Mutex
	conn        *net.UDPConn
	ControlURL  string
	ServiceType string
}

func NewUPnPClient() (*UPnPClient, error) {
	srcAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", srcAddr)
	if err != nil {
		return nil, err
	}
	return &UPnPClient{
		conn: conn,
	}, nil
}

func (u *UPnPClient) Close() error {
	u.connLock.Lock()
	defer u.connLock.Unlock()
	return u.conn.Close()
}

func (u *UPnPClient) Discover() bool {
	resps, err := u.search()
	if err != nil {
		return false
	}
	locations := u.getLocations(resps)
	fmt.Println(locations)
	for _, location := range locations {
		rootDevice, err := u.getXmlReponse(location)
		if err != nil {
			continue
		}
		controlURL, serviceType := u.findWANIPConnectionControlURL(rootDevice)
		if len(controlURL) > 0 {
			baseLocaton := location[:strings.LastIndex(location, "/")]
			u.ControlURL = baseLocaton + controlURL
			u.ServiceType = serviceType
			fmt.Printf("get control url %v for servcie type %v\n", u.ControlURL, serviceType)
			return true
		}
	}
	return false
}

func (u *UPnPClient) findWANIPConnectionControlURL(rootDevice *RootDevice) (string, string) {
	if !strings.Contains(rootDevice.Device.DeviceType, DEVICE_TYPE_ISG) {
		return "", ""
	}
	for _, device := range rootDevice.Device.DeviceList.Devices {
		if !strings.Contains(device.DeviceType, DEVICE_TYPE_WD) {
			continue
		}
		for _, subDevice := range device.DeviceList.Devices {
			if !strings.Contains(subDevice.DeviceType, DEVICE_TYPE_WCD) {
				continue
			}
			for _, service := range subDevice.ServiceList.Services {
				if strings.Contains(service.ServiceType, ST_WIPC) {
					return service.ControlURL, service.ServiceType
				} else if strings.Contains(service.ServiceType, ST_WPPPC) {
					return service.ControlURL, service.ServiceType
				}
			}
		}
	}
	return "", ""
}

func (u *UPnPClient) getXmlReponse(location string) (*RootDevice, error) {
	resp, err := http.Get(location)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	rootDevice := RootDevice{}
	err = xml.Unmarshal(body, &rootDevice)
	if err != nil {
		return nil, err
	}
	return &rootDevice, nil
}

func (u *UPnPClient) getLocations(resps []*http.Response) []string {
	locations := make([]string, 0)
	for _, resp := range resps {
		if resp.StatusCode != 200 {
			fmt.Printf("upnp get err reponse code %v\n", resp.StatusCode)
			continue
		}
		if st := resp.Header.Get("ST"); st != SEARCH_TARGET {
			fmt.Printf("upnp st not match")
			continue
		}
		location, err := resp.Location()
		if err != nil {
			fmt.Printf("upnp no location found: %v\n", err)
			continue
		}
		locationStr := strings.ToLower(location.String())
		isExist := false
		for _, exist := range locations {
			if exist == locationStr {
				isExist = true
				break
			}
		}
		if !isExist {
			locations = append(locations, locationStr)
		}
	}
	return locations
}

func (u *UPnPClient) buildSearchRequest() (http.Request, []byte, error) {
	req := http.Request{
		Header: http.Header{
			"HOST": []string{SSDP_BOARDCAST_ADDR},
			"MX":   []string{strconv.Itoa(MAX_WAIT_TIME)},
			"MAN":  []string{SSDP_DISCOVER},
			"ST":   []string{SEARCH_TARGET},
		},
	}

	var buff bytes.Buffer
	if _, err := fmt.Fprintf(&buff, "M-SEARCH * HTTP/1.1\r\n"); err != nil {
		return req, nil, err
	}
	if err := req.Header.Write(&buff); err != nil {
		return req, nil, err
	}
	if _, err := buff.Write([]byte("\r\n")); err != nil {
		return req, nil, err
	}
	return req, buff.Bytes(), nil
}

/*
* search upnp devices
*
* return:
*	-[]string upnp devlices list
 */
func (u *UPnPClient) search() ([]*http.Response, error) {
	u.connLock.Lock()
	defer u.connLock.Unlock()

	//get request info
	req, bte, err := u.buildSearchRequest()
	if err != nil {
		return nil, err
	}

	//set udp timeout
	waitTime := MAX_WAIT_TIME
	timeout := time.Duration(waitTime)*time.Second + 100*time.Millisecond
	if err := u.conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	dstAddr, err := net.ResolveUDPAddr("udp", BOARDCAST_ADDR)
	if err != nil {
		return nil, err
	}
	//send search udp
	retry := MAX_RETRY
	for i := 0; i < retry; i++ {
		n, err := u.conn.WriteTo(bte, dstAddr)
		if err != nil {
			return nil, err
		} else if n < len(bte) {
			return nil, fmt.Errorf("wrote bytes %v must be bigger than request bytess %v", n, len(bte))
		}
		time.Sleep(10 * time.Millisecond)
	}

	//get response
	resultData := make([]byte, 2048)
	resps := make([]*http.Response, 0)
	for {
		n, _, err := u.conn.ReadFrom(resultData)
		if err != nil {
			if err, ok := err.(net.Error); ok {
				if err.Timeout() {
					break
				}
				if err.Temporary() {
					time.Sleep(10 * time.Millisecond)
					continue
				}
			}
			return nil, err
		}
		resp, err := http.ReadResponse(bufio.NewReader(bytes.NewBuffer(resultData[:n])), &req)
		if err != nil {
			fmt.Printf("error processing response: %v", err)
			return nil, err
		}
		resps = append(resps, resp)
	}

	return resps, nil
}
