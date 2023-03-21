package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	ctrl "sigs.k8s.io/controller-runtime"

	"microsoft.com/azure-spring-discovery/api/logging"
	"microsoft.com/azure-spring-discovery/api/v1alpha1"
	"microsoft.com/azure-spring-discovery/core"
)

func main() {
	var server string
	var port int
	var username string
	var password string
	flag.StringVar(&server, "server", "", "Target server to be discovered")
	flag.StringVar(&username, "username", "", "Username for ssh login")
	flag.StringVar(&password, "password", "", "Password for ssh login")
	flag.IntVar(&port, "port", 0, "The ssh port")
	flag.Parse()
	cfg := &zap.Config{
		Encoding:         "console",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{"discovery.log"},
		ErrorOutputPaths: []string{"discovery.log"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,

			EncodeDuration: zapcore.MillisDurationEncoder,
		},
	}
	logger, _ := cfg.Build()
	ctrl.SetLogger(zapr.NewLogger(logger))

	ctx := context.Background()
	azureLogger := logging.GetAzureLogger(ctx)
	var serverObj = v1alpha1.SpringBootServer{
		Spec: v1alpha1.SpringBootServerSpec{
			Server: server,
			Port:   port,
		},
	}

	var executor = core.NewCoreExecutor(
		core.NewUsernamePasswordCredentialProvider(username, password),
		core.DefaultServerFactory(),
	)

	apps, err := executor.Discover(azureLogger.IntoContext(ctx), &serverObj)
	if err != nil {
		azureLogger.Error(err, "Failed to discover apps")
	}

	var specs []v1alpha1.SpringBootAppSpec
	for _, app := range apps {
		specs = append(specs, app.Spec)
	}

	if specs == nil {
		fmt.Println("Error during discovery and please check discovery.log for details")
		return
	}

	b, err := json.Marshal(specs)
	if err != nil {
		panic(err)
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(out.String())
}
