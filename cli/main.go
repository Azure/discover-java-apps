package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Azure/discover-java-apps/springboot"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	var server string
	var port int
	var username string
	var password string
	var filename string
	var format string
	flag.StringVar(&server, "server", "", "Target server to be discovered")
	flag.StringVar(&username, "username", "", "Username for ssh login")
	flag.StringVar(&password, "password", "", "Password for ssh login")
	flag.IntVar(&port, "port", 22, "The ssh port, default 22")

	flag.StringVar(&filename, "file", "", "File name for result, default console")
	flag.StringVar(&format, "format", "json", "Output format, default json")
	flag.Parse()
	cfg := &zap.Config{
		Encoding:         "console",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{"discovery.log"},
		ErrorOutputPaths: []string{"discovery.log"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			TimeKey:        "time",
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			CallerKey:      "caller",
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeDuration: zapcore.MillisDurationEncoder,
		},
	}
	logger, _ := cfg.Build()
	ctx := logr.NewContext(context.Background(), zapr.NewLogger(logger))
	azureLogger := springboot.GetAzureLogger(ctx, map[string]string{
		"server": server,
	})

	var serverConnectInfo = springboot.ServerConnectionInfo{
		Server: server,
		Port:   port,
	}

	output, err := NewOutput(filename, format)
	if err != nil {
		azureLogger.Error(err, "error when creating output", "filename", filename)
		os.Exit(1)
	}

	DoSpringBootDiscovery(ctx, serverConnectInfo, NewUsernamePasswordCredentialProvider(username, password), output)
}

func DoSpringBootDiscovery(ctx context.Context, info springboot.ServerConnectionInfo, credentialProvider springboot.CredentialProvider, output *Output) {
	azureLogger := springboot.GetAzureLogger(ctx)
	var executor = springboot.NewSpringBootDiscoveryExecutor(
		credentialProvider,
		springboot.DefaultServerConnectorFactory(
			springboot.WithConnectionTimeout(time.Duration(5)*time.Second),
			springboot.WithHostKeyCallback(MemoryHostKeyCallbackFunction()),
		),
		springboot.YamlCfg,
	)

	apps, err := executor.Discover(ctx, info)
	if err != nil {
		azureLogger.Error(err, "failed to discover")
		fmt.Println("Error occurred during discovery, please check discovery.log, any issue could report to https://github.com/Azure/azure-discovery-java-apps/issues")
		os.Exit(1)
	}

	if len(apps) == 0 {
		fmt.Print("no app discovered from " + info.Server)
		os.Exit(0)
	}

	var converter = NewSpringBootAppConverter()
	var cliApps = converter.Convert(apps)

	if err = output.Write(cliApps); err != nil {
		azureLogger.Error(err, "error when write to target file")
		fmt.Println("Error occurred while writing to file, please check discovery.log, any issue could report to https://github.com/Azure/azure-discovery-java-apps/issues")
		os.Exit(1)
	}
}
