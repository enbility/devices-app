package app

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/enbility/eebus-go/api"
	"github.com/enbility/eebus-go/service"
	shipapi "github.com/enbility/ship-go/api"
	"github.com/enbility/ship-go/cert"
	"github.com/enbility/spine-go/model"
)

const (
	eebusPort int = 4815
)

type Cem struct {
	mux         sync.Mutex
	servicesMux sync.Mutex

	eebusService *service.Service

	currentRemoteServices []shipapi.RemoteService

	servicesList []*ServiceItem

	discoveryData map[string]string
	usecaseData   map[string]string

	connections map[*Connection]bool

	allowRemoteConnections bool
}

func NewHems() *Cem {
	hems := &Cem{
		connections:            make(map[*Connection]bool),
		servicesList:           make([]*ServiceItem, 0),
		discoveryData:          make(map[string]string),
		usecaseData:            make(map[string]string),
		allowRemoteConnections: true,
	}

	return hems
}

func (c *Cem) createCertificate(certPath, keyPath string) (tls.Certificate, error) {
	certificate, err := cert.CreateCertificate("Demo", "Demo", "DE", "Demo-Unit-01")
	if err != nil {
		return tls.Certificate{}, err
	}

	pemdata := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificate.Certificate[0],
	})
	if err := os.WriteFile(certPath, pemdata, 0666); err != nil {
		log.Fatal(err)
	}

	b, err := x509.MarshalECPrivateKey(certificate.PrivateKey.(*ecdsa.PrivateKey))
	if err != nil {
		return tls.Certificate{}, err
	}
	pemdata = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	if err := os.WriteFile(keyPath, pemdata, 0666); err != nil {
		log.Fatal(err)
	}

	return certificate, nil
}

func (c *Cem) Run() {
	var err error
	var certificate tls.Certificate

	// check if there is a certificate in the working directory
	execPath := os.Args[0]
	execDir := filepath.Dir(execPath)

	certPath := execDir + "/cert.crt"
	keyPath := execDir + "/cert.key"

	portEEBUS := eebusPort
	if len(os.Args) > 2 {
		if tempPort, err := strconv.Atoi(os.Args[2]); err == nil {
			portEEBUS = tempPort
		}
	}
	log.Println("EEBUS Service running at port", portEEBUS)

	if len(os.Args) > 3 {
		certPath = os.Args[3]
	}
	if len(os.Args) > 4 {
		keyPath = os.Args[4]
	}
	certificate, err = tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		certificate, err = c.createCertificate(certPath, keyPath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("Using certificate file", certPath, "and key file", keyPath)
	}

	serial := "123456789"
	if len(os.Args) > 5 {
		serial = os.Args[5]
	}
	eebusConfiguration, err := api.NewConfiguration(
		"EnbilityNet", "EnbilityNet", "Devices-App",
		serial, model.DeviceTypeTypeEnergyManagementSystem,
		[]model.EntityTypeType{model.EntityTypeTypeCEM},
		portEEBUS, certificate, time.Second*4)
	if err != nil {
		log.Fatal(err)
	}

	c.eebusService = service.NewService(eebusConfiguration, c)
	c.eebusService.SetLogging(c)

	if err = c.eebusService.Setup(); err != nil {
		log.Fatal(err)
		return
	}

	c.eebusService.Start()
}

func (c *Cem) AddConnection(conn *Connection) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.connections[conn] = true
	c.sendAllowRemote(conn)
}

func (c *Cem) RemoveConnection(conn *Connection) {
	c.mux.Lock()
	defer c.mux.Unlock()

	delete(c.connections, conn)
	close(conn.sendChannel)
}

func (c *Cem) handleMessage(conn *Connection, message []byte) {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		return
	}

	// all are requests
	switch msg.Name {
	case MessageNameServicesList:
		c.sendServicesList(conn)
	case MessageNamePair:
		c.eebusService.RegisterRemoteSKI(msg.Ski)
	case MessageNameUnpair:
		c.eebusService.UnregisterRemoteSKI(msg.Ski)
	case MessageNameAbort:
		c.eebusService.CancelPairingWithSKI(msg.Ski)
	case MessageNameAllowRemote:
		c.allowRemoteConnections = msg.Enable
		c.broadcastAllowRemote()
	}
}

func (c *Cem) sendAllowRemote(conn *Connection) {
	msg := Message{
		Name:   MessageNameAllowRemote,
		Enable: c.allowRemoteConnections,
	}

	conn.sendMessage(msg)
}

func (c *Cem) broadcastAllowRemote() {
	c.mux.Lock()
	defer c.mux.Unlock()

	for conn := range c.connections {
		c.sendAllowRemote(conn)
	}
}

func (c *Cem) sendServicesList(conn *Connection) {
	c.updateServicesList()

	msg := Message{
		Name:     MessageNameServicesList,
		Services: c.servicesList,
	}

	conn.sendMessage(msg)
}

func (c *Cem) broadcastServicesList() {
	c.mux.Lock()
	defer c.mux.Unlock()

	for conn := range c.connections {
		c.sendServicesList(conn)
	}
}

// combine the mDNS entries with the service itself and all paired services
func (c *Cem) updateServicesList() {
	servicesList := make([]*ServiceItem, 0)

	// add the local service first
	localService := &ServiceItem{
		Ski:    c.eebusService.LocalService().SKI(),
		Brand:  "Enbility.net",
		Model:  "Devices App",
		Itself: true,
	}
	servicesList = append(servicesList, localService)

	// add the mDNS records
	for _, element := range c.currentRemoteServices {
		service := c.eebusService.RemoteServiceForSKI(element.Ski)
		detail := service.ConnectionStateDetail()

		stateError := ""
		if detail.Error() != nil {
			stateError = detail.Error().Error()
		}

		newService := &ServiceItem{
			Ski:        element.Ski,
			Trusted:    service.Trusted(),
			State:      detail.State(),
			StateError: stateError,
			Brand:      element.Brand,
			Model:      element.Model,
		}

		if service.Trusted() {
			if data, ok := c.discoveryData[element.Ski]; ok {
				newService.Discovery = data
			}

			if data, ok := c.usecaseData[element.Ski]; ok {
				newService.UseCase = data
			}
		}

		servicesList = append(servicesList, newService)
	}

	// get all paired services and add them if they are not listed
	// some services stop publishing the mDNS entry when they are connected with a paired service
	// (as they do not support multiple connected services)
	// TODO

	// sort the entries by brand, model
	sort.Slice(servicesList, func(i, j int) bool {
		item1 := servicesList[i]
		item2 := servicesList[j]
		a := strings.ToLower(item1.Brand + item1.Model + item1.Ski)
		b := strings.ToLower(item2.Brand + item2.Model + item2.Ski)
		return a < b
	})

	c.servicesMux.Lock()
	c.servicesList = servicesList
	c.servicesMux.Unlock()
}

// EEBUSServiceHandler

func (c *Cem) RemoteSKIConnected(service api.ServiceInterface, ski string) {}

func (c *Cem) RemoteSKIDisconnected(service api.ServiceInterface, ski string) {
	c.updateServicesList()

	c.broadcastServicesList()
}

func (c *Cem) VisibleRemoteServicesUpdated(service api.ServiceInterface, entries []shipapi.RemoteService) {
	c.currentRemoteServices = entries

	c.updateServicesList()

	c.broadcastServicesList()
}

func (c *Cem) ServiceShipIDUpdate(ski string, shipdID string) {}

func (c *Cem) ServicePairingDetailUpdate(ski string, detail *shipapi.ConnectionStateDetail) {
	// if accepted from both ends, we need to persist this
	if detail.State() == shipapi.ConnectionStateTrusted {
		c.eebusService.RegisterRemoteSKI(ski)
	}

	c.updateServicesList()

	c.broadcastServicesList()
}

// providing trust is only possible when there is a web interface connected
func (c *Cem) AllowWaitingForTrust(ski string) bool {
	return c.allowRemoteConnections
}

// Logging interface

func (c *Cem) Trace(args ...interface{}) {
	c.print("TRACE", args...)
}

func (c *Cem) Tracef(format string, args ...interface{}) {
	c.printFormat("TRACE", format, args...)
}

func (c *Cem) Debug(args ...interface{}) {
	c.print("DEBUG", args...)
}

func (c *Cem) Debugf(format string, args ...interface{}) {
	c.printFormat("DEBUG", format, args...)
}

func (c *Cem) Info(args ...interface{}) {
	c.print("INFO ", args...)
}

func (c *Cem) Infof(format string, args ...interface{}) {
	c.printFormat("INFO ", format, args...)
}

func (c *Cem) Error(args ...interface{}) {
	c.print("ERROR", args...)
}

func (c *Cem) Errorf(format string, args ...interface{}) {
	c.printFormat("ERROR", format, args...)
}

func (c *Cem) currentTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func (c *Cem) print(msgType string, args ...interface{}) {
	value := fmt.Sprintln(args...)
	c.filterSpineLogs(value)
	fmt.Printf("%s %s %s", c.currentTimestamp(), msgType, value)
}

func (c *Cem) printFormat(msgType, format string, args ...interface{}) {
	value := fmt.Sprintf(format, args...)
	c.filterSpineLogs(value)
	fmt.Println(c.currentTimestamp(), msgType, value)
}

// filter out UseCase and DetailedDiscovery
func (c *Cem) filterSpineLogs(msg string) {
	parts := strings.Split(msg, " ")
	if len(parts) < 3 {
		return
	}

	if parts[0] != "Recv:" {
		return
	}

	ski := parts[1]
	var msgService shipapi.RemoteService

	c.servicesMux.Lock()
	for _, service := range c.currentRemoteServices {
		if service.Ski == ski {
			msgService = service
			break
		}
	}
	c.servicesMux.Unlock()

	if len(msgService.Ski) == 0 {
		return
	}
	payload := strings.Join(parts[2:], " ")
	// discovery data
	if strings.Contains(msg, "{\"payload\":[{\"cmd\":[[{\"nodeManagementDetailedDiscoveryData\":") &&
		!strings.Contains(msg, "{\"nodeManagementDetailedDiscoveryData\":[]") {
		c.discoveryData[ski] = payload
		c.broadcastServicesList()
		return
	}

	// usecase data
	if strings.Contains(msg, "{\"payload\":[{\"cmd\":[[{\"nodeManagementUseCaseData\":") &&
		!strings.Contains(msg, "{\"nodeManagementUseCaseData\":[]") {
		c.usecaseData[ski] = payload
		c.broadcastServicesList()
		return
	}
}
