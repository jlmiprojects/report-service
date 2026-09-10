package main

import (
	"fmt"
	"log/slog"

	"blueassetgroup.com/reports-service/handlers"
	"blueassetgroup.com/reports-service/reportapi"
	"blueassetgroup.com/reports-service/repository"
	"blueassetgroup.com/reports-service/service"
	utils "blueassetgroup.com/reports-service/shared"
	"github.com/nats-io/nats.go"
)

// ServiceVersion is set at build time via -ldflags "-X main.ServiceVersion=...".
var ServiceVersion string

func main() {

	if len(ServiceVersion) == 0 {
		ServiceVersion = "0.0.0"
	}

	// CONFIG IS READ FROM conf/application.json (viper)
	config, err := utils.NewConfig()
	if err != nil {
		panic(err)
	}

	if config.Logger == nil {
		config.Logger = &utils.LoggerConfig{Type: "json", Level: "DEBUG"}
	}
	if config.Nats == nil {
		config.Nats = &utils.NatsConfig{Uri: "nats://localhost:4222"}
	}

	utils.SetupLogging(ServiceVersion, config.Logger)
	slog.Info("Logger initialized", "type", config.Logger.Type, "level", config.Logger.Level)

	nc, err := nats.Connect(config.Nats.Uri, nats.Name("report-client"), nats.Token(config.Nats.Token))
	if err != nil {
		panic(err)
	}

	repo, err := repository.NewMongo(config)
	if err != nil {
		panic(err)
	}

	// Named MongoDB datasources report scripts can query (opened lazily).
	conns := repository.NewMongoConnections(config)
	slog.Info("Configured mongo report connections", "names", conns.Names())

	serviceCalls, err := service.NewServiceCall(nc)
	if err != nil {
		panic(err)
	}

	// Wires the shared Mongo connection registry / NATS caller that every
	// report script's reportapi.Context reaches through (ctx.Mongo/ctx.Nats).
	reportapi.Init(conns, serviceCalls)

	funcMap := handlers.ReportFuncMap()
	scripts := handlers.NewScriptRunner(*config.ScriptDir)
	reportHandler := handlers.NewReportHandler(config, funcMap, scripts)

	if _, err := handlers.NewHandler(ServiceVersion, nc, config, repo, reportHandler); err != nil {
		panic(err)
	}

	e := handlers.NewServer(config, repo, funcMap, reportHandler)

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	slog.Info("Starting report http server", "addr", addr)
	e.Logger.Fatal(e.Start(addr))
}
