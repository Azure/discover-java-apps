package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Azure/discover-java-apps/weblogic"
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
	var weblogicusername string
	var weblogicpassword string
	var weblogicport int
	flag.StringVar(&server, "server", "", "Target server to be discovered")
	flag.StringVar(&username, "username", "", "Username for ssh login")
	flag.StringVar(&password, "password", "", "Password for ssh login")
	flag.IntVar(&port, "port", 22, "The ssh port, default 22")
	flag.StringVar(&weblogicusername, "weblogicusername", "weblogic", "Username for weblogic login")
	flag.StringVar(&weblogicpassword, "weblogicpassword", "", "Password for weblogic login")
	flag.IntVar(&weblogicport, "weblogicport", 7001, "The weblogic port, default 7001")

	flag.StringVar(&filename, "file", "", "File name for result, default console")
	flag.StringVar(&format, "format", "json", "Output format, default json")
	flag.Parse()
	cfg := &zap.Config{
		Encoding:         "console",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{"weblogic.log"},
		ErrorOutputPaths: []string{"weblogic.log"},
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
	azureLogger := weblogic.GetAzureLogger(ctx, map[string]string{
		"server": server,
	})

	output, err := NewOutput(filename, format)
	if err != nil {
		azureLogger.Error(err, "error when creating output", "filename", filename)
		os.Exit(1)
	}

	var weblogicServerConnectInfo = weblogic.ServerConnectionInfo{
		Server: server,
		Port:   port,
	}

	DoWebLogicDiscovery(ctx, weblogicServerConnectInfo, WebLogicNewUsernamePasswordCredentialProvider(username, password, weblogicusername, weblogicpassword, weblogicport), output)
}

func DoWebLogicDiscovery(ctx context.Context, info weblogic.ServerConnectionInfo, credentialProvider weblogic.CredentialProvider, output *Output) {
	println("Discovering weblogic apps ::start ------------------------------------------------------")
	azureLogger := weblogic.GetAzureLogger(ctx)
	var executor = weblogic.NewWeblogicDiscoveryExecutor(
		credentialProvider,
		weblogic.DefaultServerConnectorFactory(
			weblogic.WithConnectionTimeout(time.Duration(5)*time.Second),
			weblogic.WithHostKeyCallback(MemoryHostKeyCallbackFunction()),
		),
	)

	apps, err := executor.Discover(ctx, info)

	if err != nil {
		azureLogger.Error(err, "failed to discover")
		fmt.Println("Error occurred during discovery, please check weblogic.log, any issue could report to https://github.com/Azure/azure-discovery-java-apps/issues")
		os.Exit(1)
	}

	if len(apps) == 0 {
		fmt.Print("no weblogic app discovered from " + info.Server)
		os.Exit(0)
	}
	// this is used to append the result
	var converter = NewWeblogicAppConverter()
	var weblogicCliApp = converter.Convert(apps)

	if err = output.Write(weblogicCliApp); err != nil {
		azureLogger.Error(err, "error when write to target file")
		fmt.Println("Error occurred while writing to file, please check weblogic.log, any issue could report to https://github.com/Azure/azure-discovery-java-apps/issues")
		os.Exit(1)
	}
	println("\n\nDiscovering weblogic apps ::end ------------------------------------------------------")

}
