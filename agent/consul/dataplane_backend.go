package consul

import (
	"github.com/hashicorp/consul/acl"
	"github.com/hashicorp/consul/agent/grpc/dataplane"
)

type dataplaneBackend struct {
	srv *Server
}

func (s dataplaneBackend) ResolveToken(token string) (acl.Authorizer, error) {
	return s.srv.ResolveToken(token)
}

var _ dataplane.Backend = (*dataplaneBackend)(nil)
