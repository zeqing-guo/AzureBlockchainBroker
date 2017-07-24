package main

import (
	"os"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagerflags"
	// "github.com/Azure/go-autorest/autorest"
	// "github.com/Azure/go-autorest/autorest/adal"
	"github.com/zeqing-guo/AzureBlockchainBroker/azureblockchainbroker"
)

func main() {
	sink, err := lager.NewRedactingWriterSink(os.Stdout, lager.INFO, nil, nil)
	if err != nil {
		panic(err)
	}
	logger, logsink := lagerflags.NewFromSink("azureblockchainbroker", sink)
	logger.Info("starting")
	defer logger.Info("end")
}
