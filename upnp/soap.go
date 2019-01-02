package upnp

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/symphonyprotocol/log"
)

var soapLogger = log.GetLogger("soap")

type ExternalIPAddress struct {
	IPAddress string `xml:"NewExternalIPAddress"`
}

type GetGenericPortMappingEntryResponse struct {
	NewProtocol       string `xml:"NewProtocol"`
	NewInternalClient string `xml:"NewInternalClient"`
	NewPort           int    `xml:"NewExternalPort"`
}

type SOAPNode struct {
	Name       string
	Value      string
	Attributes map[string]string
	Children   []*SOAPNode
}

func NewSOAPNode(name string, val string) *SOAPNode {
	return &SOAPNode{
		Name:  name,
		Value: val,
	}
}

func (s *SOAPNode) Add(node *SOAPNode) {
	s.Children = append(s.Children, node)
}

func (s *SOAPNode) ToString() string {
	buffer := bytes.NewBufferString("<")
	buffer.WriteString(s.Name)
	for key, val := range s.Attributes {
		buffer.WriteString(" ")
		buffer.WriteString(key)
		buffer.WriteString("=")
		buffer.WriteString(val)
	}
	buffer.WriteString(">")
	buffer.WriteString(s.Value)
	for _, child := range s.Children {
		buffer.WriteString(child.ToString())
	}
	buffer.WriteString("</")
	buffer.WriteString(s.Name)
	buffer.WriteString(">")
	return buffer.String()
}

func AddPortMapping(localAddr string, localPort int, remotePort int, protocol string, u *UPnPClient) bool {
	soap := &SOAPNode{
		Name: "m:AddPortMapping",
		Attributes: map[string]string{
			"xmlns:m": "\"" + u.ServiceType + "\"",
		},
	}
	soap.Add(NewSOAPNode("NewRemoteHost", ""))
	soap.Add(NewSOAPNode("NewExternalPort", strconv.Itoa(remotePort)))
	soap.Add(NewSOAPNode("NewProtocol", protocol))
	soap.Add(NewSOAPNode("NewInternalPort", strconv.Itoa(localPort)))
	soap.Add(NewSOAPNode("NewInternalClient", localAddr))
	soap.Add(NewSOAPNode("NewEnabled", "1"))
	soap.Add(NewSOAPNode("NewPortMappingDescription", "symphonyprotocol"))
	soap.Add(NewSOAPNode("NewLeaseDuration", "0"))
	_, err := soapRequest(u.ControlURL, soap, "AddPortMapping", u.ServiceType)
	if err != nil {
		soapLogger.Error("%v", err)
		return false
	}
	return true
}

func DeletePortMapping(remotePort int, protocol string, u *UPnPClient) bool {
	soap := &SOAPNode{
		Name: "m:DeletePortMapping",
		Attributes: map[string]string{
			"xmlns:m": "\"" + u.ServiceType + "\"",
		},
	}
	soap.Add(NewSOAPNode("NewRemoteHost", ""))
	soap.Add(NewSOAPNode("NewExternalPort", strconv.Itoa(remotePort)))
	soap.Add(NewSOAPNode("NewProtocol", protocol))
	_, err := soapRequest(u.ControlURL, soap, "DeletePortMapping", u.ServiceType)
	if err != nil {
		soapLogger.Error("%v", err)
		return false
	}
	return true
}

func GetExternalIPAddress(u *UPnPClient) (string, error) {
	soap := &SOAPNode{
		Name: "u:GetExternalIPAddress",
		Attributes: map[string]string{
			"xmlns:u=": "\"" + u.ServiceType + "\"",
		},
	}
	resp, err := soapRequest(u.ControlURL, soap, "GetExternalIPAddress", u.ServiceType)
	if err != nil {
		return "", err
	}
	externalIP := ExternalIPAddress{}
	reader := strings.NewReader(resp)
	decoder := xml.NewDecoder(reader)
	for {
		token, err := decoder.Token()
		if err != nil || token == nil || token == io.EOF {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			inElement := t.Name.Local
			if inElement == "GetExternalIPAddressResponse" {
				decoder.DecodeElement(&externalIP, &t)
				break
			}
		default:
		}
	}
	if externalIP.IPAddress == "" {
		return "", fmt.Errorf("cannot find external IP")
	}
	return externalIP.IPAddress, nil
}

func GetGenericPortMappingEntry(index int, u *UPnPClient) (string, string, int, error) {
	soap := &SOAPNode{
		Name: "u:GetGenericPortMappingEntry",
		Attributes: map[string]string{
			"xmlns:u=": "\"" + u.ServiceType + "\"",
		},
	}
	soap.Add(NewSOAPNode("NewPortMappingIndex", strconv.Itoa(index)))
	resp, err := soapRequest(u.ControlURL, soap, "GetGenericPortMappingEntry", u.ServiceType)
	if err != nil {
		soapLogger.Error("%v", err)
		return "", "", 0, err
	}
	reader := strings.NewReader(resp)
	decoder := xml.NewDecoder(reader)
	extPort := GetGenericPortMappingEntryResponse{}
	for {
		token, err := decoder.Token()
		if err != nil || token == nil || token == io.EOF {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			inElement := t.Name.Local
			if inElement == "GetGenericPortMappingEntryResponse" {
				decoder.DecodeElement(&extPort, &t)
				break
			}
		}
	}
	return extPort.NewProtocol, extPort.NewInternalClient, extPort.NewPort, nil
}

func soapRequest(urlStr string, soap *SOAPNode, function string, serviceType string) (string, error) {
	envelope := &SOAPNode{
		Name: "s:Envelope",
		Attributes: map[string]string{
			"xmlns:s":         "\"http://schemas.xmlsoap.org/soap/envelope/\"",
			"s:encodingStyle": "\"http://schemas.xmlsoap.org/soap/encoding/\"",
		},
	}
	body := NewSOAPNode("s:Body", "")
	body.Add(soap)
	envelope.Add(body)
	sendText := "<?xml version=\"1.0\"?>" + envelope.ToString()
	//fmt.Printf("SOAP post: %v\n", sendText)
	bodyBytes := []byte(sendText)
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	req := &http.Request{
		Method: "POST",
		URL:    u,
		Header: http.Header{
			"SOAPACTION":   []string{"\"" + serviceType + "#" + function + "\""},
			"CONTENT-TYPE": []string{"text/xml; charset=\"utf-8\""},
		},
		Body:          ioutil.NopCloser(bytes.NewBuffer(bodyBytes)),
		ContentLength: int64(len(bodyBytes)),
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("soap request get:%v", resp.Status)
	}
	bte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	//fmt.Printf("SOAP recieve: %v", string(bte))
	return string(bte), nil
}
