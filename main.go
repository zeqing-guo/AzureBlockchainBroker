package main

import (
	"flag"
	"fmt"
	"os"

	"code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagerflags"
	"github.com/pivotal-cf/brokerapi"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/zeqing-guo/AzureBlockchainBroker/broker"
	"github.com/zeqing-guo/AzureBlockchainBroker/utils"
)

// Broker
var atAddress = flag.String(
	"listenAddr",
	"0.0.0.0:9000",
	"(optional) - host:port to serve service broker API",
)

var serviceName = flag.String(
	"serviceName",
	"azureblockchain",
	"(optional) - name of the service to register with cloud controller",
)

var serviceID = flag.String(
	"serviceID",
	"abb90071-f3e2-4a31-99f0-fc5d552dbbba",
	"(optional) - ID of the service to register with cloud controller",
)

// Azure
var environment = flag.String(
	"environment",
	"AzureCloud",
	"[REQUIRED] - The environment for Azure Management Service. AzureCloud, AzureChinaCloud, AzureUSGovernment, AzureGermanCloud or AzureStack.",
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
	"(optional) - Administrator username of each deployed VM (alphanumeric characters only)",
)

var adminPassword = flag.String(
	"adminPassword",
	"",
	"[REQUIRED] - Administrator password for each deployed VM",
)

var ethereumAccountPsswd = flag.String(
	"ethereumAccountPsswd",
	"",
	"[REQUIRED] - Password used to secure the default Ethereum account that will be generated",
)

var ethereumAccountPassphrase = flag.String(
	"ethereumAccountPassphrase",
	"",
	"[REQUIRED] - Password used to generate the private key associated with the default Ethereum account that is generated.  Consider a password with sufficient randomness to ensure a strong private key",
)

var ethereumNetworkID = flag.Uint64(
	"ethereumNetworkID",
	553289,
	"(optional) - Private Ethereum network ID to which to connect (max 9 digit number)",
)

var numConsortiumMembers = flag.Uint64(
	"numConsortiumMembers",
	2,
	"(optional) - Number of members within the network.  Each member's nodes live in their own subnet.",
)

var numMiningNodesPerMember = flag.Uint64(
	"numMiningNodesPerMember",
	1,
	"(optional) - Number of mining nodes to create for each consortium member.",
)

var mnNodeVMSize = flag.String(
	"mnNodeVMSize",
	"Standard_D1_v2",
	"(optional) - Size of the virtual machine used for mining nodes",
)

var numTXNodes = flag.Uint64(
	"numTXNodes",
	1,
	"(optional) - Number of load balanced transaction nodes",
)

var txNodeVMSize = flag.String(
	"txNodeVMSize",
	"Standard_D1_v2",
	"(optional) - Size of the virtual machine for transaction nodes",
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
	vmSizes := []string{
		"Standard_A1",
		"Standard_A2",
		"Standard_A3",
		"Standard_A4",
		"Standard_A5",
		"Standard_A6",
		"Standard_A7",
		"Standard_D1",
		"Standard_D2",
		"Standard_D3",
		"Standard_D4",
		"Standard_D11",
		"Standard_D12",
		"Standard_D13",
		"Standard_D14",
		"Standard_D1_v2",
		"Standard_D2_v2",
		"Standard_D3_v2",
		"Standard_D4_v2",
		"Standard_D5_v2",
		"Standard_D11_v2",
		"Standard_D12_v2",
		"Standard_D13_v2",
		"Standard_D14_v2",
		"Standard_D15_v2",
		"Standard_F1",
		"Standard_F2",
		"Standard_F4",
		"Standard_F8",
		"Standard_F16",
	}

	if *adminPassword == "" {
		fmt.Fprint(os.Stderr, "\nError: adminPassword is required\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if *clientID == "" {
		fmt.Fprint(os.Stderr, "\nError: clientID is required\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if *clientSecret == "" {
		fmt.Fprint(os.Stderr, "\nError: clientSecret is required\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if *ethereumAccountPsswd == "" {
		fmt.Fprint(os.Stderr, "\nError: ethereumAccountPsswd is required\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if *ethereumAccountPassphrase == "" {
		fmt.Fprint(os.Stderr, "\nError: ethereumAccountPassphrase is required\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if *namePrefix == "" {
		fmt.Fprint(os.Stderr, "\nError: namePrefix is required\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if *serviceID == "" {
		fmt.Fprint(os.Stderr, "\nError: serviceID is required\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if *serviceName == "" {
		fmt.Fprint(os.Stderr, "\nError: serviceName is required\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if *subscriptionID == "" {
		fmt.Fprint(os.Stderr, "\nError: subscriptionID is required\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if *tenantID == "" {
		fmt.Fprint(os.Stderr, "\nError: tenantID is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// template parameters requirements
	if len(*namePrefix) > 6 {
		fmt.Fprint(os.Stderr, "\nnamePrefix should be 6 alphanumeric characters or less\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if len(*adminUsername) < 1 || len(*adminUsername) > 64 {
		fmt.Fprint(os.Stderr, "\nadminUsername should be not be void and 64 characters or less\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if len(*adminPassword) < 12 || len(*adminPassword) > 72 {
		fmt.Fprint(os.Stderr, "\nadminPassword should be in [12, 64]\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if len(*ethereumAccountPsswd) < 12 {
		fmt.Fprint(os.Stderr, "\nethereumAccountPsswd should be 12 alphanumeric characters or more\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if len(*ethereumAccountPassphrase) < 12 {
		fmt.Fprint(os.Stderr, "\nethereumAccountPassphrase should be 12 alphanumeric characters or more\n\n")
		flag.Usage()
		os.Exit(1)
	}
	fmt.Println(ethereumNetworkID)
	if (*ethereumNetworkID < 5) || (*ethereumNetworkID >= 2147483648) {
		fmt.Fprint(os.Stderr, "\nethereumNetworkID should be in [5, 2^31)\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if *numConsortiumMembers < 2 || *numConsortiumMembers > 5 {
		fmt.Fprint(os.Stderr, "\nnumConsortiumMembers should be in [2, 5]\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if *numMiningNodesPerMember < 1 || *numMiningNodesPerMember > 19 {
		fmt.Fprint(os.Stderr, "\nnumMiningNodesPerMember should be in [1, 19]\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if !stringInSlice(*mnNodeVMSize, vmSizes) {
		fmt.Fprint(os.Stderr, "\nUnsupported mining node VM size\n\n")
		flag.Usage()
		os.Exit(1)
	}
	if !stringInSlice(*txNodeVMSize, vmSizes) {
		fmt.Fprint(os.Stderr, "\nUnsupported transaction node VM size\n\n")
		flag.Usage()
		os.Exit(1)
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func createServer(logger lager.Logger) ifrit.Runner {
	azureConfig := broker.NewAzureConfig(
		*environment,
		*tenantID,
		*clientID,
		*clientSecret,
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
