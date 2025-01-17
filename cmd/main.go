package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/CloudDetail/apo-apm-adapter/pkg/apmtrace"
	"github.com/CloudDetail/apo-apm-adapter/pkg/config"
	"github.com/CloudDetail/apo-apm-adapter/pkg/global"
	"github.com/CloudDetail/apo-apm-adapter/pkg/httpserver"

	"github.com/spf13/viper"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "apm-adapter.yml", "Configuration file")
	flag.Parse()
	adapterCfg, err := readInConfig(*configPath)
	if err != nil {
		return fmt.Errorf("fail to read configuration: %w", err)
	}
	apmTraceClient, err := apmtrace.NewApmTraceClient(adapterCfg.TraceApi, adapterCfg.Timeout)
	if err != nil {
		return fmt.Errorf("fail to connect apm trace client: %w", err)
	}
	global.TRACE_CLIENT = apmTraceClient

	httpserver.StartHttpServer(adapterCfg.HttpPort)
	return nil
}

func readInConfig(path string) (*config.AdapterConfig, error) {
	viper := viper.New()
	viper.SetConfigFile(path)
	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		return nil, fmt.Errorf("error happened while reading config file: %w", err)
	}
	adapterCfg := &config.AdapterConfig{}
	_ = viper.UnmarshalKey("adapter", adapterCfg)

	return adapterCfg, nil
}
