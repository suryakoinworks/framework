package routes

import (
	"context"

	configs "github.com/crowdeco/bima/configs"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type MuxRouter struct {
	Routes []configs.Route
}

func (m *MuxRouter) Register(routes []configs.Route) {
	m.Routes = routes
}

func (m *MuxRouter) Handle(context context.Context, server *runtime.ServeMux, client *grpc.ClientConn) {
	for _, v := range m.Routes {
		v.SetClient(client)
		server.HandlePath(v.Method(), v.Path(), v.Handle)
	}
}

func (m *MuxRouter) Priority() int {
	return configs.LOWEST_PRIORITY - 1
}
