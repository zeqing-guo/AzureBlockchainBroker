package broker_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAzurefilebroker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BlockchainBroker Suite")
}
