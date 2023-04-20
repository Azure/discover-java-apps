package springboot

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSpringBootDiscovery(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SpringBoot discovery test suit")
}
