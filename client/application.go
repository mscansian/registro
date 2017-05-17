package client

// NewApplication return a new Application object the specified name.
func NewApplication(name string) *Application {
	return &Application{
		Name:      name,
		Instances: make([]*Instance, 0),
	}
}

// Application represents an app registered to the server.
type Application struct {
	// Name specifies a name to diferentiate apps.
	Name string `json:"name"`

	// Instances holds a list of instances running this app.
	Instances []*Instance `json:"instances"`
}

// GetInstance return the instance with the specified id.
func (a *Application) GetInstance(id string) *Instance {
	for _, inst := range a.Instances {
		if inst.Id == id {
			return inst
		}
	}
	return nil
}

// GetAvailableInstances returns all instances with status UP.
func (a *Application) GetAvailableInstances() []*Instance {
	instances := make([]*Instance, 0)
	for _, inst := range a.Instances {
		if inst.Status == UP {
			instances = append(instances, inst)
		}
	}
	return instances
}
