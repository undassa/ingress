//
// Last.Backend LLC CONFIDENTIAL
// __________________
//
// [2014] - [2018] Last.Backend LLC
// All Rights Reserved.
//
// NOTICE:  All information contained herein is, and remains
// the property of Last.Backend LLC and its suppliers,
// if any.  The intellectual and technical concepts contained
// herein are proprietary to Last.Backend LLC
// and its suppliers and may be covered by Russian Federation and Foreign Patents,
// patents in process, and are protected by trade secret or copyright law.
// Dissemination of this information or reproduction of this material
// is strictly forbidden unless prior written permission is obtained
// from Last.Backend LLC.
//

package envs

import (
	"github.com/lastbackend/lastbackend/pkg/network"
	"text/template"

	"github.com/lastbackend/ingress/pkg/ingress/state"
	"github.com/lastbackend/lastbackend/pkg/api/client/types"
)

var _env Env

type Env struct {
	net    *network.Network
	state  *state.State
	client types.IngressClientV1
	config struct {
		tpl  *template.Template
		path string
		name string
		pid  string
	}
	haproxy string
	dns     struct {
		Endpoint string
		Cluster  []string
		External []string
	}
}

func Get() *Env {
	return &_env
}

func (c *Env) SetNet(n *network.Network) {
	c.net = n
}

func (c *Env) GetNet() *network.Network {
	return c.net
}

func (c *Env) SetState(state *state.State) {
	c.state = state
}

func (c *Env) GetState() *state.State {
	return c.state
}

func (c *Env) SetClient(client types.IngressClientV1) {
	c.client = client
}

func (c *Env) GetClient() types.IngressClientV1 {
	return c.client
}

func (c *Env) SetTemplate(t *template.Template, path, name, pid string) {
	c.config.tpl = t
	c.config.path = path
	c.config.name = name
	c.config.pid = pid
}

func (c *Env) GetTemplate() (*template.Template, string, string, string) {
	return c.config.tpl, c.config.path, c.config.name, c.config.pid
}

func (c *Env) SetHaproxy(exec string) {
	c.haproxy = exec
}

func (c *Env) GetHaproxy() string {
	return c.haproxy
}
