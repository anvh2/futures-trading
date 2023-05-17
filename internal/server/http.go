package server

import (
	"context"
	"expvar"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	defaultMetricsPath = "/metrics"
	defaultDebugPath   = "/debug/vars"
)

// serve http request
func (s *Server) httpServe(ctx context.Context, l net.Listener) error {
	// configure mux options
	muxOpts := []runtime.ServeMuxOption{
		runtime.WithMarshalerOption("*", &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				UseEnumNumbers:  true,
				EmitUnpopulated: true,
			},
		}),
	}

	mux := runtime.NewServeMux(muxOpts...)

	// register handler
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	endPoint := fmt.Sprintf("localhost:%d", viper.GetInt("server.port"))

	err := s.register.http(ctx, mux, endPoint, opts)
	if err != nil {
		return err
	}

	mux.HandlePath(http.MethodGet, defaultDebugPath, func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		expvar.Handler().ServeHTTP(w, r)
	})
	mux.HandlePath(http.MethodGet, defaultMetricsPath, func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		promhttp.Handler().ServeHTTP(w, r)
	})

	// add middlewares
	var handler http.Handler
	handler = mux

	server := &http.Server{Handler: handler}

	s.server.http = server
	return server.Serve(l)
}
