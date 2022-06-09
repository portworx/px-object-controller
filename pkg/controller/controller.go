package controller

import (
	"github.com/libopenstorage/openstorage/pkg/correlation"
)

const (
	componentNameController = correlation.Component("pkg/controller")
)

var (
	logger = correlation.NewPackageLogger(componentNameController)
)

// Config represents a configuration for creating a controller server
type Config struct {
}

// Server represents a controller server
type Server struct {
	config *Config
}

// New returns a new controller server with a zuora billing client
func New(cfg *Config) (*Server, error) {
	return &Server{
		config: cfg,
	}, nil
}

// Run starts the Px Object Service controller
func (s *Server) Run(threadCount int, stopCh chan struct{}) {

}
