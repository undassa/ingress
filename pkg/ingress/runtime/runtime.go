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

package runtime

import (
	"context"
	"github.com/lastbackend/ingress/pkg/ingress/envs"
	"github.com/lastbackend/lastbackend/pkg/distribution/types"
	"github.com/lastbackend/lastbackend/pkg/log"
)

const (
	logRuntimePrefix = "ingress:runtime"
	logLevel         = 3
)

type Runtime struct {
	spec    chan *types.IngressManifest
	process *Process
}

// Restore node runtime state
func (r *Runtime) Restore(ctx context.Context) {
	log.V(logLevel).Debugf("%s:restore:> restore init", logRuntimePrefix)

	if err := envs.Get().GetNet().EndpointRestore(ctx); err != nil {
		log.Errorf("%s:> can not restore endpoins: %s", err.Error())
	}

	if err := envs.Get().GetNet().SubnetRestore(ctx); err != nil {
		log.Errorf("%s:> can not restore network: %s", err.Error())
	}

	if err := envs.Get().GetNet().ResolverManage(ctx); err != nil {
		log.Errorf("%s:> can not manage resolver:%s",logRuntimePrefix,  err.Error())
	}

	if err := configCheck(); err != nil {
		log.Errorf("can no sync config: %s", err.Error())
		return
	}

	if err := r.process.manage(); err != nil {
		log.Errorf("can not manage haproxy process: %s", err.Error())
		return
	}
}

// Sync node runtime with new spec
func (r *Runtime) Sync(ctx context.Context, spec *types.IngressManifest) error {
	log.V(logLevel).Debugf("%s:sync:> sync runtime state", logRuntimePrefix)
	r.spec <- spec
	return nil
}

func (r *Runtime) Loop(ctx context.Context) {

	log.V(logLevel).Debugf("%s:loop:> start runtime loop", logRuntimePrefix)

	go func(ctx context.Context) {
		for {
			select {
			case spec := <-r.spec:

				log.V(logLevel).Debugf("%s:loop:> provision new spec", logRuntimePrefix)

				if spec.Meta.Initial {

					log.V(logLevel).Debugf("%s> clean up endpoints", logRuntimePrefix)
					endpoints := envs.Get().GetNet().Endpoints().GetEndpoints()
					for e := range endpoints {

						if e == envs.Get().GetNet().GetResolverEndpointKey() {
							continue
						}

						if _, ok := spec.Endpoints[e]; !ok {
							envs.Get().GetNet().EndpointDestroy(context.Background(), e, endpoints[e])
						}
					}

					log.V(logLevel).Debugf("%s> clean up networks", logRuntimePrefix)
					nets := envs.Get().GetNet().Subnets().GetSubnets()

					for cidr := range nets {
						if _, ok := spec.Network[cidr]; !ok {
							envs.Get().GetNet().SubnetDestroy(ctx, cidr)
						}
					}
				}

				if len(spec.Meta.Discovery) != 0 {
					log.V(logLevel).Debugf("%s>set cluster dns ips: %#v", logRuntimePrefix, spec.Meta.Discovery)
					for key, res := range spec.Meta.Discovery {
						envs.Get().GetNet().Resolvers().SetResolver(key, res)
						envs.Get().GetNet().ResolverManage(ctx)
					}
				}

				log.V(logLevel).Debugf("%s> provision init", logRuntimePrefix)

				log.V(logLevel).Debugf("%s> provision endpoints", logRuntimePrefix)
				for e, spec := range spec.Endpoints {
					log.V(logLevel).Debugf("endpoint: %v", e)
					if err := envs.Get().GetNet().EndpointManage(ctx, e, spec); err != nil {
						log.Errorf("Endpoint [%s] manage err: %s", e, err.Error())
					}
				}

				log.V(logLevel).Debugf("%s> provision networks", logRuntimePrefix)
				for cidr, n := range spec.Network {
					log.V(logLevel).Debugf("network: %v", n)
					if err := envs.Get().GetNet().SubnetManage(ctx, cidr, n); err != nil {
						log.Errorf("Subnet [%s] create err: %s", n.CIDR, err.Error())
					}
				}

				log.V(logLevel).Debugf("%s> provision routes", logRuntimePrefix)
				for e, spec := range spec.Routes {
					log.V(logLevel).Debugf("route: %v", e)
					if err := RouteManage(ctx, e, spec); err != nil {
						log.Errorf("Route [%s] manage err: %s", e, err.Error())
						continue
					}

					r.process.reload()
				}
			}
		}
	}(ctx)
}

func NewRuntime() *Runtime {
	r := new(Runtime)
	r.spec = make(chan *types.IngressManifest)
	r.process = new(Process)
	return r
}
