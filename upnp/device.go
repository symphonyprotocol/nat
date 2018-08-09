package upnp

import (
	"encoding/xml"
)

type RootDevice struct {
	XMLName         xml.Name    `xml:"root"`
	SpecVersion     SpecVersion `xml:"specVersion"`
	Device          Device      `xml:"device"`
	PresentationURL string      `xml:"presentationURL"`
}

type SpecVersion struct {
	Major int `xml:"major"`
	Minor int `xml:"minor"`
}

type Device struct {
	DeviceType       string      `xml:"deviceType"`
	FriendlyName     string      `xml:"friendlyName"`
	Manufacturer     string      `xml:"manufacturer"`
	ManufacturerURL  string      `xml:"manufacturerURL"`
	ModelDescription string      `xml:"modelDescription"`
	ModelName        string      `xml:"modelName"`
	ModelNumber      string      `xml:"modelNumber"`
	ModelURL         string      `xml:"modelURL"`
	SerialNumber     string      `xml:"serialNumber"`
	UDN              string      `xml:"UDN"`
	UPC              string      `xml:"UPC"`
	IconList         IconList    `xml:"iconList"`
	ServiceList      ServiceList `xml:"serviceList"`
	DeviceList       DeviceList  `xml:"deviceList"`
}

type IconList struct {
	Icons []Icon `xml:"icon"`
}

type Icon struct {
	MimeType string `xml:"mimetype"`
	Width    int    `xml:"width"`
	Height   int    `xml:"height"`
	Depth    string `xml:"depth"`
	Url      string `xml:"url"`
}

type ServiceList struct {
	Services []Service `xml:"service"`
}

type Service struct {
	ServiceType string `xml:"serviceType"`
	ServiceId   string `xml:"serviceId"`
	SCPDURL     string `xml:"SCPDURL"`
	ControlURL  string `xml:"controlURL"`
	EventSubURL string `xml:"eventSubURL"`
}

type DeviceList struct {
	Devices []Device `xml:"device"`
}
