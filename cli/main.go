package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"

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
	azureLogger := logging.GetAzureLogger(context.Background())
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

	apps, err := executor.Discover(context.Background(), &serverObj)
	if err != nil {
		azureLogger.Error(err, "Failed to discover apps")
	}

	var specs []v1alpha1.SpringBootAppSpec
	for _, app := range apps {
		specs = append(specs, app.Spec)
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
