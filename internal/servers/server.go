package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	exchange_cache "github.com/anvh2/futures-trading/internal/cache/exchange"
	market_cache "github.com/anvh2/futures-trading/internal/cache/market"
	"github.com/anvh2/futures-trading/internal/config"
	"github.com/anvh2/futures-trading/internal/externals/binance"
	"github.com/anvh2/futures-trading/internal/externals/telegram"
	"github.com/anvh2/futures-trading/internal/libs/channel"
	"github.com/anvh2/futures-trading/internal/libs/logger"
	"github.com/anvh2/futures-trading/internal/libs/queue"
	"github.com/anvh2/futures-trading/internal/servers/handler"
	"github.com/anvh2/futures-trading/internal/servers/orchestrator"
	"github.com/anvh2/futures-trading/internal/services/settings"
	pb "github.com/anvh2/futures-trading/pkg/api/v1/signal"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/soheilhy/cmux"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

// RegisterGRPCHandlerFunc register server from
type RegisterGRPCHandlerFunc func(s *grpc.Server)

// RegisterHTTPHandlerFunc ...
type RegisterHTTPHandlerFunc func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)

type Server struct {
	logger *logger.Logger

	handler      *handler.Handler
	orchestrator *orchestrator.ServiceOrchestrator

	server *struct {
		grpc *grpc.Server
		http *http.Server
	}

	register *struct {
		grpc RegisterGRPCHandlerFunc
		http RegisterHTTPHandlerFunc
	}

	quitChannel chan struct{}
}

func New(config config.Config) *Server {
	logger, err := logger.New(viper.GetString("trading.log_path"))
	if err != nil {
		log.Fatal("failed to init logger", err)
	}

	notify, err := telegram.NewTelegramBot(logger, viper.GetString("telegram.token"))
	if err != nil {
		// log.Fatal("failed to new chat bot", err)
	}

	binance := binance.New(logger, false)
	marketCache := market_cache.NewMarket(viper.GetInt32("chart.candles.limit"))
	exchange := exchange_cache.New(logger)
	handler := handler.New()
	quit := make(chan struct{})

	queue := queue.New()
	channel := channel.New()
	settings := settings.NewDefaultSettings()

	// Create service orchestrator
	orchestrator, err := orchestrator.NewServiceOrchestrator(
		config, logger, binance, notify, marketCache, exchange, queue, channel, settings)
	if err != nil {
		log.Fatal("failed to create service orchestrator", err)
	}

	return &Server{
		logger: logger,

		handler:      handler,
		orchestrator: orchestrator,

		server: &struct {
			grpc *grpc.Server
			http *http.Server
		}{},

		register: &struct {
			grpc RegisterGRPCHandlerFunc
			http RegisterHTTPHandlerFunc
		}{
			grpc: func(s *grpc.Server) { pb.RegisterSignalServiceServer(s, handler) },
			http: pb.RegisterSignalServiceHandlerFromEndpoint,
		},

		quitChannel: quit,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", viper.GetInt("server.port")))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := s.orchestrator.Start(ctx); err != nil {
		log.Fatal("failed to start service orchestrator", zap.Error(err))
	}

	// catch sig
	sigs := make(chan os.Signal, 1)
	done := make(chan error, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	serverCtx, serverCancel := context.WithCancel(context.Background())

	go func() {
		sig := <-sigs
		fmt.Println("Exiting...: ", sig)

		s.server.grpc.Stop()
		s.server.http.Close()

		serverCancel()
		close(s.quitChannel)

		if err := s.orchestrator.Stop(); err != nil {
			fmt.Println("Error stopping orchestrator:", err)
		}

		close(done)
	}()

	go s.serve(serverCtx, lis)

	fmt.Println("Server now listening at: " + lis.Addr().String())

	fmt.Println("Ctrl-C to interrupt...")
	e := <-done
	fmt.Println("Shutted down.", zap.Error(e))
	return e
}

// start listening grpc & http & exporter request
func (s *Server) serve(ctx context.Context, listener net.Listener) {
	m := cmux.New(listener)
	grpcListener := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	httpListener := m.Match(cmux.HTTP1Fast())

	g := new(errgroup.Group)
	g.Go(func() error { return s.grpcServe(ctx, grpcListener) })
	g.Go(func() error { return s.httpServe(ctx, httpListener) })
	g.Go(func() error { return m.Serve() })

	g.Wait()
}
