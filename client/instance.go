package client

import (
	"time"
)

// NewInstance return a new Instance object with the specified data.
func NewInstance(id, ip string, port int) *Instance {
	return &Instance{
		Id:          id,
		IPAddr:      ip,
		Port:        port,
		Status:      STARTING,
		LastRenewal: time.Now().Unix(),
	}
}

// Instance represents a service running an application.
type Instance struct {
	// Id is a unique identifier for an Instance.
	Id string `json:"id"`

	// IPAddr is the newowrk address where the instance is located.
	IPAddr string `json:"ip"`

	// Port is the network port where the instance is located.
	Port int `json:"port"`

	// Status provide information of the operational status of the instance.
	Status StatusType `json:"status"`

	// LastRenewal holds the timestamp when the instance last contacted the SR.
	LastRenewal int64 `json:"lastRenewal"`
}

// StatusType represents an instance status
type StatusType string

const (
	// UP represents an instance receiving requests.
	UP StatusType = "up"

	// DOWN representes an instance that has not sent heartbeats after some time.
	DOWN StatusType = "down"

	// STARTING represents an instance that has registered, but has not yet send any heartbeats.
	STARTING StatusType = "starting"

	// OUTOFSERVICE represents an instance that has been deliberately deleted.
	// It may be down for maintainance or shutting down.
	OUTOFSERVICE StatusType = "out-of-service"
)
