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
	"text/template"

	"github.com/lastbackend/ingress/pkg/ingress/state"
	"github.com/lastbackend/lastbackend/pkg/api/client/types"
	"github.com/lastbackend/lastbackend/pkg/runtime/cni"
	"github.com/lastbackend/lastbackend/pkg/runtime/cpi"
)

var _env Env

type Env struct {
	cni    cni.CNI
	cpi    cpi.CPI
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

func (c *Env) SetCNI(n cni.CNI) {
	c.cni = n
}

func (c *Env) GetCNI() cni.CNI {
	return c.cni
}

func (c *Env) SetCPI(n cpi.CPI) {
	c.cpi = n
}

func (c *Env) GetCPI() cpi.CPI {
	return c.cpi
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

func (c *Env) SetDNSResolver(ip string) {
	c.dns.Endpoint = ip
}

func (c *Env) GetDNSResolver() string {
	return c.dns.Endpoint
}

func (c *Env) SetClusterDNS(dns []string) {
	c.dns.Cluster = dns
}

func (c *Env) GetClusterDNS() []string {
	return c.dns.Cluster
}

func (c *Env) SetExternalDNS(dns []string) {

	if len(dns) == 0 {
		c.dns.External = []string{"8.8.8.8", "8.8.4.4"}
	}
	c.dns.External = dns
}

func (c *Env) GetExternalDNS() []string {
	return c.dns.External
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
