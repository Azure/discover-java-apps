module github.com/Azure/discover-java-apps

go 1.20

replace github.com/Azure/discover-java-apps/springboot => ./springboot

require (
	github.com/Azure/discover-java-apps/springboot v0.0.0-00010101000000-000000000000
	github.com/go-logr/logr v1.2.4
	github.com/go-logr/zapr v1.2.3
	go.uber.org/zap v1.24.0
	golang.org/x/crypto v0.8.0
)

require (
	github.com/creekorful/mvnparser v1.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dprotaso/go-yit v0.0.0-20191028211022-135eb7262960 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/onsi/gomega v1.27.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.13.5 // indirect
	github.com/vmware-labs/yaml-jsonpath v0.3.2 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/mod v0.10.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
