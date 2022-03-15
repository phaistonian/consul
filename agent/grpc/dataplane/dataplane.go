package dataplane

import (
	"github.com/hashicorp/consul/acl"
	"github.com/hashicorp/consul/proto-public/pbdataplane"
	"github.com/hashicorp/go-hclog"
)

type Server struct {
	Backend Backend
	Logger  Logger
}

func NewServer(backend Backend, logger Logger) *Server {
	return &Server{Backend: backend, Logger: logger}
}

type Logger interface {
	Trace(msg string, args ...interface{})
	With(args ...interface{}) hclog.Logger
}

var _ pbdataplane.DataplaneServiceServer = (*Server)(nil)

type Backend interface {
	ResolveToken(token string) (acl.Authorizer, error)
}
