package app

import (
	shipapi "github.com/enbility/ship-go/api"
)

type ServiceItem struct {
	Ski        string                  `json:"ski"`       // the services ski
	Trusted    bool                    `json:"trusted"`   // if the service is already trusted
	State      shipapi.ConnectionState `json:"state"`     // the connection state
	StateError string                  `json:"error"`     // the connection error message if in error state
	Brand      string                  `json:"brand"`     // the services brand string
	Model      string                  `json:"model"`     // the services model string
	Itself     bool                    `json:"itself"`    // Defines if this entry is about this service itself
	Discovery  string                  `json:"discovery"` // The SPINE json string
	UseCase    string                  `json:"usecase"`   // The SPINE json string
}

type Message struct {
	Name     MessageName    `json:"name"`     // The message type
	Ski      string         `json:"ski"`      // The SKI the message is meant for, if applicable
	Services []*ServiceItem `json:"services"` // The services list, if applicable
	// Service  ServiceItem    `json:"service"`  // The service item, if applicable
	Enable bool `json:"enable"` // Used with MessageNameAllowRemote
}

type MessageName string

const (
	MessageNameService      MessageName = "service"      // Request or send specific ServiceItem, requires SKI or service
	MessageNameServicesList MessageName = "serviceslist" // Request or send ServicesList, requires Services
	MessageNamePair         MessageName = "pair"         // Request or send pairing request, requires SKI
	MessageNameAbort        MessageName = "abort"        // Abort or deny the pairing process
	MessageNameUnpair       MessageName = "unpair"       // Request unpairing request, requires SKI
	MessageNameDiscovery    MessageName = "discovery"    // Send discovery data
	MessageNameUsecase      MessageName = "usecase"      // Send usecase data
	MessageNameAllowRemote  MessageName = "allowremote"  // Enable/Disable allowing remote connections for trust
)
