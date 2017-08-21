package main

import (
	"flag"
	// "fmt"
	"os"
	// "strconv"

	"code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagerflags"
	"github.com/pivotal-cf/brokerapi"
	// "github.com/Azure/go-autorest/autorest"
	// "github.com/Azure/go-autorest/autorest/adal"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/zeqing-guo/AzureBlockchainBroker/broker"
	"github.com/zeqing-guo/AzureBlockchainBroker/utils"
)

// Broker
var atAddress = flag.String(
	"listenAddr",
	"0.0.0.0:$PORT",
	"host:port to serve service broker API",
)

var serviceName = flag.String(
	"serviceName",
	"azureblockchain",
	"name of the service to register with cloud controller",
)

var serviceID = flag.String(
	"serviceID",
	"06948cb0-cad7-4buh-leba-9ed8b5c345a3",
	"ID of the service to register with cloud controller",
)

// Azure
var environment = flag.String(
	"environment",
	"AzureCloud",
	"The environment for Azure Management Service. AzureCloud, AzureChinaCloud, AzureUSGovernment, AzureGermanCloud or AzureStack.",
)

var tenantID = flag.String(
	"tenantID",
	"",
	"[REQUIRED] - The tenant id for your service principal.",
)

var clientID = flag.String(
	"clientID",
	"",
	"[REQUIRED] - The client id for your service principal.",
)

var clientSecret = flag.String(
	"clientSecret",
	"",
	"[REQUIRED] - The client secret for your service principal.",
)

var azureStackDomain = flag.String(
	"azureStackDomain",
	"",
	"Required when environment is AzureStack. The domain for your AzureStack deployment.",
)

var azureStackAuthentication = flag.String(
	"azureStackAuthentication",
	"",
	"Required when environment is AzureStack. The authentication type for your AzureStack deployment. AzureAD, AzureStackAD or AzureStack.",
)

var azureStackResource = flag.String(
	"azureStackResource",
	"",
	"Required when environment is AzureStack. The token resource for your AzureStack deployment.",
)

var azureStackEndpointPrefix = flag.String(
	"azureStackEndpointPrefix",
	"",
	"Required when environment is AzureStack. The endpoint prefix for your AzureStack deployment.",
)

// Resource deployment
var subscriptionID = flag.String(
	"subscriptionID",
	"",
	"[REQUIRED] - The subscription ID for resource.",
)

var location = flag.String(
	"location",
	"southcentralus",
	"[REQUIRED] - The location for deploying template",
)

// Blockchain configuration
var namePrefix = flag.String(
	"namePrefix",
	"blockchain",
	"[REQUIRED] - String used as a base for naming resources (6 alphanumeric characters or less).  A unique hash is prepended to the string for some resources, while resource-specific information is appended.",
)

var adminUsername = flag.String(
	"adminUsername",
	"gethadmin",
	"[REQUIRED] - Administrator username of each deployed VM (alphanumeric characters only)",
)

var adminPassword = flag.String(
	"adminPassword",
	"",
	"[REQUIRED] - Administrator password for each deployed VM",
)

var ethereumAccountPsswd = flag.String(
	"ethereumAccountPsswd",
	"",
	"Password used to secure the default Ethereum account that will be generated",
)

var ethereumAccountPassphrase = flag.String(
	"ethereumAccountPassphrase",
	"",
	"Password used to generate the private key associated with the default Ethereum account that is generated.  Consider a password with sufficient randomness to ensure a strong private key",
)

var ethereumNetworkID = flag.Uint64(
	"ethereumNetworkID",
	72,
	"Private Ethereum network ID to which to connect (max 9 digit number)",
)

var numConsortiumMembers = flag.Uint64(
	"numConsortiumMembers",
	2,
	"Number of members within the network.  Each member's nodes live in their own subnet.",
)

var numMiningNodesPerMember = flag.Uint64(
	"numMiningNodesPerMember",
	1,
	"Number of mining nodes to create for each consortium member.",
)

var mnNodeVMSize = flag.String(
	"mnNodeVMSize",
	"Standard_D1_v2",
	"Size of the virtual machine used for mining nodes",
)

var numTXNodes = flag.Uint64(
	"numTXNodes",
	1,
	"Number of load balanced transaction nodes",
)

var txNodeVMSize = flag.String(
	"txNodeVMSize",
	"Standard_D1_v2",
	"Size of the virtual machine for transaction nodes",
)

var (
	username string
	password string
)

func main() {
	parseCommandLine()
	parseEnvironment()

	checkParams()

	sink, err := lager.NewRedactingWriterSink(os.Stdout, lager.INFO, nil, nil)
	if err != nil {
		panic(err)
	}
	logger, logSink := lagerflags.NewFromSink("azureblockchainbroker", sink)
	logger.Info("start")
	defer logger.Info("end")

	server := createServer(logger)

	if dbgAddr := debugserver.DebugAddress(flag.CommandLine); dbgAddr != "" {
		server = utils.ProcessRunnerFor(grouper.Members{
			{"debug-server", debugserver.Runner(dbgAddr, logSink)},
			{"broker-api", server},
		})
	}

	process := ifrit.Invoke(server)
	logger.Info("started")
	utils.UntilTerminated(logger, process)
}

func parseCommandLine() {
	lagerflags.AddFlags(flag.CommandLine)
	debugserver.AddFlags(flag.CommandLine)
	flag.Parse()
}

func parseEnvironment() {
	username, _ = os.LookupEnv("USERNAME")
	password, _ = os.LookupEnv("PASSWORD")
}

func checkParams() {

}

func createServer(logger lager.Logger) ifrit.Runner {
	azureConfig := broker.NewAzureConfig(
		*environment,
		*tenantID,
		*clientID,
		*clientSecret,
		"",
		"",
	)
	azureStackConfig := broker.NewAzureStackConfig(*azureStackDomain, *azureStackAuthentication, *azureStackResource, *azureStackEndpointPrefix)
	cloudConfig := broker.NewCloudConfig(*azureConfig, *azureStackConfig)

	resourceConfig := broker.NewResourceConfig(
		*subscriptionID,
		"",
		true,
		*location,
		"",
		false,
		false,
	)

	blockchainConfig := broker.NewBlockchainConfig(
		*namePrefix,
		*adminUsername,
		*adminPassword,
		*ethereumAccountPsswd,
		*ethereumAccountPassphrase,
		*ethereumNetworkID,
		*numConsortiumMembers,
		*numMiningNodesPerMember,
		*mnNodeVMSize,
		*numTXNodes,
		*txNodeVMSize,
	)

	credentials := brokerapi.BrokerCredentials{Username: username, Password: password}
	serviceBroker, err := broker.New(
		logger,
		*cloudConfig,
		*resourceConfig,
		*blockchainConfig,
		*serviceName,
		*serviceID,
	)
	if err != nil {
		panic(err)
	}
	handler := brokerapi.New(serviceBroker, logger.Session("broker-api"), credentials)

	return http_server.New(*atAddress, handler)
}
