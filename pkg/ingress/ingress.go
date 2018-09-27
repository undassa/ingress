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

package ingress

import (
	"context"
	"github.com/lastbackend/ingress/pkg/ingress/controller"
	"github.com/lastbackend/ingress/pkg/ingress/envs"
	"github.com/lastbackend/ingress/pkg/ingress/runtime"
	"github.com/lastbackend/ingress/pkg/ingress/state"
	"github.com/lastbackend/lastbackend/pkg/api/client"
	"github.com/lastbackend/lastbackend/pkg/distribution/types"
	"github.com/lastbackend/lastbackend/pkg/log"
	"github.com/lastbackend/lastbackend/pkg/runtime/cni/cni"
	"github.com/lastbackend/lastbackend/pkg/runtime/cpi/cpi"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
	"text/template"
)



func Daemon() bool {

	var (
		sigs = make(chan os.Signal)
		done = make(chan bool, 1)
	)

	log.New(viper.GetInt("verbose"))
	log.Info("Start Ingress server")

	_cni, err := cni.New()
	if err != nil {
		log.Errorf("Cannot initialize cni: %v", err)
	}

	_cpi, err := cpi.New()
	if err != nil {
		log.Errorf("Cannot initialize cni: %v", err)
	}

	envs.Get().SetCNI(_cni)
	envs.Get().SetCPI(_cpi)

	ip := viper.GetString("ingress.resolver.ip")
	if ip == types.EmptyString {
		ip = "172.17.0.1"
	}
	envs.Get().SetDNSResolver(ip)
	envs.Get().SetExternalDNS(viper.GetStringSlice("dns.ips"))

	st := state.New()

	st.Ingress().Info = runtime.IngressInfo()
	st.Ingress().Status = runtime.IngressStatus()

	envs.Get().SetState(st)
	envs.Get().SetTemplate(template.Must(template.New("").Parse(runtime.HaproxyTemplate)),
		viper.GetString("haproxy.path"),
		viper.GetString("haproxy.name"),
		viper.GetString("haproxy.pid"))

	envs.Get().SetHaproxy(viper.GetString("haproxy.exec"))

	r := runtime.NewRuntime()

	if viper.IsSet("api") || viper.IsSet("api_uri") {

		cfg := client.NewConfig()
		cfg.BearerToken = viper.GetString("token")

		if viper.IsSet("api.tls") && !viper.GetBool("api.tls.insecure") {
			cfg.TLS = client.NewTLSConfig()
			cfg.TLS.CertFile = viper.GetString("api.tls.cert")
			cfg.TLS.KeyFile = viper.GetString("api.tls.key")
			cfg.TLS.CAFile = viper.GetString("api.tls.ca")
		}

		endpoint := viper.GetString("api.uri")
		if viper.IsSet("api_uri") {
			endpoint = viper.GetString("api_uri")
		}

		rest, err := client.New(client.ClientHTTP, endpoint, cfg)
		if err != nil {
			log.Errorf("Init client err: %s", err)
		}

		c := rest.V1().Ingress(st.Ingress().Info.Hostname)
		envs.Get().SetClient(c)
		ctl := controller.New(r)

		if err := ctl.Connect(context.Background()); err != nil {
			log.Errorf("ingress:initialize: connect err %s", err.Error())
		}

		go ctl.Sync(context.Background())
	}

	go func() {
		types.SecretAccessToken = viper.GetString("token")
		r.Restore(context.Background())
		r.Loop(context.Background())
	}()

	// Handle SIGINT and SIGTERM.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-sigs:
				done <- true
				return
			}
		}
	}()

	<-done

	log.Info("Handle SIGINT and SIGTERM.")

	return true
}
