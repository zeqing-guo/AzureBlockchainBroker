package main

import (
	"flag"
	// "fmt"
	"os"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagerflags"
	// "github.com/Azure/go-autorest/autorest"
	// "github.com/Azure/go-autorest/autorest/adal"
	"github.com/zeqing-guo/AzureBlockchainBroker/broker"
)

func main() {
	parseCommandLine()
	sink, err := lager.NewRedactingWriterSink(os.Stdout, lager.INFO, nil, nil)
	if err != nil {
		panic(err)
	}
	logger, _ := lagerflags.NewFromSink("azureblockchainbroker", sink)
	logger.Info("start")
	defer logger.Info("end")

	azureConfig := broker.NewAzureConfig(
		"",
		"",
		"",
		"",
		"",
		"",
	)
	azureStackConfig := broker.NewAzureStackConfig("", "", "", "")
	cloudConfig := broker.NewCloudConfig(*azureConfig, *azureStackConfig)
	configuration := broker.NewConfiguration(
		"c4528d9e-c99a-48bb-b12d-fde2176a43b8",
		"binxi-zeqing-testsb1",
		true,
		"southcentralus",
		"",
		false,
		false,
	)
	blockchainConfig := broker.NewBlockchainConfiguration(
		"ethnet",
		"",
		"",
		"",
		"",
		23219,
		2,
		1,
		"Standard_A1",
		1,
		"Standard_A1",
	)
	account, err := broker.NewAzureAccount(logger, *cloudConfig, *configuration, *blockchainConfig)
	if err != nil {
		panic(err)
	}
	// err = account.Create()
	// res := <-account.DeploymentResult
	// logger.Info("deployment-response-data", lager.Data{
	// 	"Parameters":     res.Properties.Parameters,
	// 	"ParametersLink": res.Properties.ParametersLink,
	// })
	err = account.Create()
	if err != nil {
		panic(err)
	}
	// logger.Info("Deployment-Get", lager.Data{
	// "de": de,
	// })
}

func parseCommandLine() {
	lagerflags.AddFlags(flag.CommandLine)
}
