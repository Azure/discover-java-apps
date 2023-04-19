package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Azure/discover-java-apps/springboot"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"io"
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
	var serverConnectInfo = springboot.ServerConnectionInfo{
		Server: server,
		Port:   port,
	}

	azuerLogger := springboot.GetAzureLogger(ctx, map[string]string{
		"server": server,
	})
	var executor = springboot.NewSpringBootDiscoveryExecutor(
		NewUsernamePasswordCredentialProvider(username, password),
		springboot.DefaultServerConnectorFactory(
			springboot.WithConnectionTimeout(time.Duration(5)*time.Second),
			springboot.WithHostKeyCallback(MemoryHostKeyCallbackFunction()),
		),
		springboot.YamlCfg,
	)

	apps, err := executor.Discover(ctx, serverConnectInfo)
	if err != nil {
		azuerLogger.Error(err, "failed to discover")
		os.Exit(1)
	}

	if len(apps) == 0 {
		azuerLogger.Warning(fmt.Errorf("no app discovered"), "")
		os.Exit(0)
	}

	var converter = NewSpringBootAppConverter()
	var cliApps = converter.Convert(apps)

	var writer io.Writer
	if writer, err = NewWriter(filename); err != nil {
		panic(err)
	}
	if closer, ok := writer.(io.Closer); ok {
		defer closer.Close()
	}

	output := NewOutput[*CliApp](writer, format)
	if err = output.Write(cliApps); err != nil {
		azuerLogger.Error(err, "error when write to target file", "filename", filename)
		os.Exit(1)
	}
}
