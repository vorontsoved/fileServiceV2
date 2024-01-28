package grpcapp

import (
	"fileservice/internal/service/config"
	fileSystemGRPC "fileservice/internal/transport/grpc/fileservice"
	"fmt"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"net"
)

const op = "grpcapp.Run"

type App struct {
	log        zerolog.Logger
	gRPCServer *grpc.Server
	cfg        config.Config
}

func New(log zerolog.Logger, fileService fileSystemGRPC.FileService, cfg config.Config) (*App, error) {
	gRPCServer := grpc.NewServer()
	fileSystemGRPC.RegisterServerAPI(gRPCServer, fileService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		cfg:        cfg,
	}, nil
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	log := a.log.With().Str("op", op).Int("port", a.cfg.GRPC.Port).Logger()

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.GRPC.Port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info().Str("addr", l.Addr().String()).Msg("grpc server is running")

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	log := a.log.With().Str("op", op).Logger()

	log.Info().Int("port", a.cfg.GRPC.Port).Msg("stopping gRPC server")

	a.gRPCServer.GracefulStop()
}
