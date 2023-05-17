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

	"github.com/anvh2/futures-trading/internal/cache"
	"github.com/anvh2/futures-trading/internal/cache/exchange"
	"github.com/anvh2/futures-trading/internal/cache/market"
	"github.com/anvh2/futures-trading/internal/channel"
	"github.com/anvh2/futures-trading/internal/libs/queue"
	"github.com/anvh2/futures-trading/internal/logger"
	"github.com/anvh2/futures-trading/internal/server/analyzer"
	"github.com/anvh2/futures-trading/internal/server/crawler"
	"github.com/anvh2/futures-trading/internal/server/handler"
	"github.com/anvh2/futures-trading/internal/server/orderer"
	"github.com/anvh2/futures-trading/internal/services/binance"
	"github.com/anvh2/futures-trading/internal/services/telegram"
	"github.com/anvh2/futures-trading/internal/settings"
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
	logger  *logger.Logger
	binance *binance.Binance
	notify  *telegram.TelegramBot

	queue    *queue.Queue
	channel  *channel.Channel
	settings *settings.Settings

	marketCache   cache.Market
	exchangeCache cache.Exchange

	crawler  *crawler.Crawler
	analyzer *analyzer.Analyzer
	orderer  *orderer.Orderer
	handler  *handler.Handler

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

func New() *Server {
	logger, err := logger.New(viper.GetString("trading.log_path"))
	if err != nil {
		log.Fatal("failed to init logger", err)
	}

	notify, err := telegram.NewTelegramBot(logger, viper.GetString("telegram.token"))
	if err != nil {
		log.Fatal("failed to new chat bot", err)
	}

	binance := binance.New(logger, false)
	market := market.NewMarket(viper.GetInt32("chart.candles.limit"))
	exchange := exchange.New(logger)
	handler := handler.New()
	quit := make(chan struct{})

	queue := queue.New()
	channel := channel.New()
	settings := settings.NewDefaultSettings()

	return &Server{
		logger:  logger,
		binance: binance,
		notify:  notify,

		queue:    queue,
		channel:  channel,
		settings: settings,

		marketCache:   market,
		exchangeCache: exchange,

		crawler:  crawler.New(logger, binance, notify, market, exchange, channel),
		analyzer: analyzer.New(logger, notify, market, exchange, queue, channel, settings),
		orderer:  orderer.New(logger, notify, market, exchange, queue, settings),
		handler:  handler,

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
	if err := s.crawler.Start(); err != nil {
		log.Fatal("failed to crawling data", zap.Error(err))
	}

	if err := s.analyzer.Start(); err != nil {
		log.Fatal("failed to start analyzer", zap.Error(err))
	}

	if err := s.orderer.Start(); err != nil {
		log.Fatal("failed to start orderer", zap.Error(err))
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", viper.GetInt("server.port")))
	if err != nil {
		return err
	}

	// catch sig
	sigs := make(chan os.Signal, 1)
	done := make(chan error, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sig := <-sigs
		fmt.Println("Exiting...: ", sig)

		close(s.quitChannel)

		s.orderer.Stop()
		s.analyzer.Stop()
		s.crawler.Stop()

		s.server.grpc.Stop()
		s.server.http.Close()

		cancel()
		close(done)
	}()

	go s.serve(ctx, lis)

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
