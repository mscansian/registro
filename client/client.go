// Package client provide functions to send requests to the Service Registry REST server.
//
// Example:
//     c := client.NewClient("http://localhost:8080/registro")
//     app, inst, err := c.RegisterService("service-id", "app-name", "127.0.0.1", 8080)
//     if err != nil {
//       log.Fatal(err)
//     }
//     go func() {
//       for {
//         err := c.RenewInstance(app, inst)
// 	       if err != nil {
//           log.Printf("service renew error: %s", err)
//         }
//         <-time.After(30 * time.Second)
//       }
//     }()
//     <-time.After(90 * time.Second)
package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// NewClient returns a Client with the specified ServiceUrl.
func NewClient(url string) *Client {
	return &Client{
		ServiceUrl: url,
	}
}

// Client represents a connection to the SR REST server.
type Client struct {
	// Root URL to SR server.
	ServiceUrl string
}

// RegisterService register the application and an instance to the SR.
func (c *Client) RegisterService(id, appName, ip string, port int) (*Application, *Instance, error) {
	app, err := c.GetApp(appName)
	if err != nil {
		log.Printf("app %s not found. registering.", appName)
		if app, err = c.NewApp(appName); err != nil {
			return nil, nil, err
		}
	}

	inst, err := c.NewInstance(app, id, ip, port)
	if err != nil {
		return nil, nil, err
	}
	return app, inst, nil
}

// GetApps makes a request to SR and return the list of registered apps.
func (c *Client) GetApps() ([]*Application, error) {
	body, err := c.get("/apps", 200)
	if err != nil {
		return nil, err
	}

	var r struct {
		Apps []*Application `json:"applications"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}

	return r.Apps, nil
}

// GetApps makes a request to SR and return the Application with the specified name.
func (c *Client) GetApp(name string) (*Application, error) {
	apps, err := c.GetApps()
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		if app.Name == name {
			return app, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Application %s not found.", name))
}

// UpdateApplication makes a request to SR and update the app list of Instances.
func (c *Client) UpdateApplication(app *Application) error {
	body, err := c.get("/apps/"+app.Name, 200)
	if err != nil {
		return err
	}

	var r struct {
		Instances []*Instance `json:"instances"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return err
	}
	app.Instances = r.Instances
	return nil
}

// NewApp makes a request to SR and create a new Application.
func (c *Client) NewApp(name string) (*Application, error) {
	app := NewApplication(name)
	r, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		return nil, err
	}

	_, err = c.post("/apps", r, 201)
	if err != nil {
		return nil, err
	}
	return app, nil
}

// NewInstance makes a request to SR and create a new app Instance.
func (c *Client) NewInstance(app *Application, id, ip string, port int) (*Instance, error) {
	inst := NewInstance(id, ip, port)
	r, err := json.MarshalIndent(inst, "", "  ")
	if err != nil {
		return nil, err
	}

	_, err = c.post("/apps/"+app.Name, r, 201)
	if err != nil {
		return nil, err
	}
	return inst, nil
}

// RenewInstance makes a request to SR and update Instance heartbeat.
func (c *Client) RenewInstance(app *Application, inst *Instance) error {
	_, err := c.do(http.MethodPut, "/apps/"+app.Name+"/"+inst.Id, 204)
	if err != nil {
		return err
	}
	return nil
}

// DeleteInstance makes a request to SR and delete instance.
func (c *Client) DeleteInstance(app *Application, inst *Instance) error {
	_, err := c.do(http.MethodDelete, "/apps/"+app.Name+"/"+inst.Id, 204)
	if err != nil {
		return err
	}
	return nil
}

// get makes a GET request to the SR.
func (c *Client) get(url string, expectedCode int) ([]byte, error) {
	r, err := http.Get(c.ServiceUrl + "/1.0" + url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != expectedCode {
		return nil, &UnexpectedCodeError{Code: r.StatusCode}
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// post makes a POST request to the SR.
func (c *Client) post(url string, postdata []byte, expectedCode int) ([]byte, error) {
	buf := bytes.NewBuffer(postdata)
	r, err := http.Post(c.ServiceUrl+"/1.0"+url, "application/json", buf)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != expectedCode {
		return nil, &UnexpectedCodeError{Code: r.StatusCode}
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// do makes an HTTP request to the SR with the specified method.
func (c *Client) do(method, url string, expectedCode int) ([]byte, error) {
	cli := &http.Client{}
	req, err := http.NewRequest(method, c.ServiceUrl+"/1.0"+url, nil)
	if err != nil {
		return nil, err
	}

	r, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != expectedCode {
		return nil, &UnexpectedCodeError{Code: r.StatusCode}
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
