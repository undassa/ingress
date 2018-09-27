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
	logResolverPrefix = "runtime:resolver:"
	resolverEndpointKey = "resolver"
)

func ResolverManage (ctx context.Context) error {

	log.V(logLevel).Debugf("%s:> create resolver", logResolverPrefix)

	manifest := generateResolverManifest()

	endpointState := envs.Get().GetState().Endpoints().GetEndpoint(resolverEndpointKey)
	if endpointState != nil {
		if endpointEqual(manifest, endpointState) {
			return nil
		}

		state, err := EndpointUpdate(ctx, resolverEndpointKey, endpointState, manifest)
		if err != nil {
			log.Errorf("%s:> can not update endpoint", logResolverPrefix)
			return err
		}

		envs.Get().GetState().Endpoints().SetEndpoint(resolverEndpointKey, state)
		return nil
	}


	state, err := EndpointCreate(ctx, resolverEndpointKey, manifest)
	if err != nil {
		log.Errorf("%s:> can not create endpoint", logResolverPrefix)
		return err
	}

	envs.Get().GetState().Endpoints().SetEndpoint(resolverEndpointKey, state)
	return nil
}


func generateResolverManifest () *types.EndpointManifest {

	manifest := new(types.EndpointManifest)
	manifest.IP = envs.Get().GetDNSResolver()

	cdns := envs.Get().GetClusterDNS()
	edns := envs.Get().GetExternalDNS()

	if len(cdns) == 0 {
		manifest.Upstreams = edns
	} else {
		manifest.Upstreams = cdns
	}

	manifest.PortMap = make(map[uint16]string)
	manifest.PortMap[53]="53/udp"

	return manifest
}