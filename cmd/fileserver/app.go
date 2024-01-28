package main

import (
	"context"
	"fileservice/internal/app"
	"fileservice/internal/service/config"
	"fileservice/internal/service/storage/postgres"
	grpcapp "fileservice/internal/transport/grpc"
	"fmt"
	"github.com/jackc/pgx/v5"
	"io"
	"os"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"gopkg.in/natefinch/lumberjack.v2"
)

type App struct {
	isWinService bool
	grpc         *grpc.Server
	db           *pgx.Conn
}

func New(isWinService bool) *App {
	return &App{isWinService: isWinService}
}

func (a *App) Run() error {
	cfg, err := config.New("")
	if err != nil {
		return fmt.Errorf("can't create new config: %w", err)
	}

	var logger zerolog.Logger

	LogFileWritter := &lumberjack.Logger{
		Filename:   cfg.Logging.Filename,
		MaxSize:    cfg.Logging.MaxSize,
		MaxAge:     cfg.Logging.MaxAge,
		MaxBackups: cfg.Logging.MaxBackups,
		LocalTime:  cfg.Logging.LocalTime,
		Compress:   true,
	}

	if a.isWinService {
		logger = zerolog.New(LogFileWritter).With().Timestamp().Logger()
	} else {
		mw := io.MultiWriter(os.Stdout, LogFileWritter)

		logger = zerolog.New(mw).With().Timestamp().Logger()
	}

	db, err := postgres.InitDatabase(logger, cfg)
	if err != nil {
		return err
	}

	service, err := app.New(logger, db)
	if err != nil {
		return err
	}

	grpc, err := grpcapp.New(logger, service, cfg)
	if err != nil {
		return err
	}

	grpc.MustRun()

	return nil
}

func (a *App) Close() error {

	a.grpc.Stop()

	a.db.Close(context.Background())
	return nil
}
