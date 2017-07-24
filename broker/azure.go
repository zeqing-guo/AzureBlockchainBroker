package broker

import (
	"encoding/json"
	"fmt"
	"strconv"

	"code.cloudfoundry.org/lager"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/to"
)

const (
	creator                     = "creator"
	resourceNotFound            = "StatusCode=404"
	fileRequestTimeoutInSeconds = 60
	locationWestUS              = "westus"
	deploymentName              = "blockchainDeployment"
)

var (
	userAgent = "azureblockchainbroker"
	// encryptionKeySource = "Microsoft.Storage"
)

const (
	AzureCloud        = "AzureCloud"
	AzureChinaCloud   = "AzureChinaCloud"
	AzureGermanCloud  = "AzureGermanCloud"
	AzureUSGovernment = "AzureUSGovernment"
	AzureStack        = "AzureStack"
)

type BlockchainConfiguration struct {
	NamePrefix                string `json:"namePrefix"`
	AdminUsername             string `json:"adminUsername"`
	AdminPassword             string `json:"adminPassword"`
	EthereumAccountPsswd      string `json:"ethereumAccountPsswd"`
	EthereumAccountPassphrase string `json:"ethereumAccountPassphrase"`
	EthereumNetworkID         string `json:"ethereumNetworkID"`
	NumConsortiumMembers      string `json:"numConsortiumMembers"`
	NumMiningNodesPerMember   string `json:"numMiningNodesPerMember"`
	MnNodeVMSize              string `json:"mnNodeVMSize"`
	NumTXNodes                string `json:"numTXNodes"`
	TxNodeVMSize              string `json:"txNodeVMSize"`
}

type APIVersions struct {
	Storage         string
	Group           string
	ActiveDirectory string
}

type Environment struct {
	ResourceManagerEndpointURL string
	ActiveDirectoryEndpointURL string
	APIVersions                APIVersions
}

var Environments = map[string]Environment{
	AzureCloud: Environment{
		ResourceManagerEndpointURL: "https://management.azure.com/",
		ActiveDirectoryEndpointURL: "https://login.microsoftonline.com",
		APIVersions: APIVersions{
			Storage:         "2016-05-31",
			Group:           "2016-06-01",
			ActiveDirectory: "2015-06-15",
		},
	},
	AzureChinaCloud: Environment{
		ResourceManagerEndpointURL: "https://management.chinacloudapi.cn/",
		ActiveDirectoryEndpointURL: "https://login.chinacloudapi.cn",
		APIVersions: APIVersions{
			Storage:         "2015-06-15",
			Group:           "2016-06-01",
			ActiveDirectory: "2015-06-15",
		},
	},
	AzureUSGovernment: Environment{
		ResourceManagerEndpointURL: "https://management.usgovcloudapi.net/",
		ActiveDirectoryEndpointURL: "https://login.microsoftonline.com",
		APIVersions: APIVersions{
			Storage:         "2015-06-15",
			Group:           "2016-06-01",
			ActiveDirectory: "2015-06-15",
		},
	},
	AzureGermanCloud: Environment{
		ResourceManagerEndpointURL: "https://management.microsoftazure.de/",
		ActiveDirectoryEndpointURL: "https://login.microsoftonline.de",
		APIVersions: APIVersions{
			Storage:         "2015-06-15",
			Group:           "2016-06-01",
			ActiveDirectory: "2015-06-15",
		},
	},
	AzureStack: Environment{
		APIVersions: APIVersions{
			Storage:         "2015-06-15",
			Group:           "2016-06-01",
			ActiveDirectory: "2015-06-15",
		},
	},
}

type AzureAccount struct {
	logger                    lager.Logger
	cloudConfig               *CloudConfig
	SubscriptionID            string
	ResourceGroupName         string
	UseHTTPS                  bool
	Location                  string
	CustomDomainName          string
	UseSubDomain              bool
	EnableEncryption          bool
	baseURL                   string
	resourcesManagementClient *resources.DeploymentsClient
	blockchainconfiguration   *BlockchainConfiguration
}

func NewAzureAccount(logger lager.Logger, cloudConfig *CloudConfig, configuration Configuration) (*AzureAccount, error) {
	logger = logger.Session("azure-account", lager.Data{
		"CloudConfig": cloudConfig,
	})
	azureAccount := AzureAccount{
		logger:                    logger,
		cloudConfig:               cloudConfig,
		SubscriptionID:            configuration.SubscriptionID,
		ResourceGroupName:         configuration.ResourceGroupName,
		UseHTTPS:                  true,
		Location:                  locationWestUS,
		CustomDomainName:          "",
		UseSubDomain:              false,
		EnableEncryption:          false,
		baseURL:                   "",
		resourcesManagementClient: nil,
		blockchainconfiguration:   nil,
	}
	if configuration.UseHTTPS != "" {
		if ret, err := strconv.ParseBool(configuration.UseHTTPS); err == nil {
			azureAccount.UseHTTPS = ret
		}
	}
	if configuration.Location != "" {
		azureAccount.Location = configuration.Location
	}
	if configuration.UseSubDomain != "" {
		if ret, err := strconv.ParseBool(configuration.UseSubDomain); err == nil {
			azureAccount.UseSubDomain = ret
		}
	}
	if configuration.EnableEncryption != "" {
		if ret, err := strconv.ParseBool(configuration.EnableEncryption); err == nil {
			azureAccount.EnableEncryption = ret
		}
	}

	if err := azureAccount.initManagementClient(); err != nil {
		return nil, err
	}
	return &azureAccount, nil
}

func (account *AzureAccount) initManagementClient() error {
	logger := account.logger.Session("init-management-client")
	logger.Info("start")
	defer logger.Info("end", nil)

	environment := account.cloudConfig.Azure.Environment
	tenantID := account.cloudConfig.Azure.TenanID
	clientID := account.cloudConfig.Azure.ClientID
	clientSecret := account.cloudConfig.Azure.ClientSecret
	oauthConfig, err := adal.NewOAuthConfig(Environments[environment].ActiveDirectoryEndpointURL, tenantID)
	if err != nil {
		logger.Error("new-oauth-config", err, lager.Data{
			"Environment":                environment,
			"ActiveDirectoryEndpointURL": Environments[environment].ActiveDirectoryEndpointURL,
			"TenanID":                    tenantID,
		})
		return fmt.Errorf("Error in initManagementClient: %v", err)
	}

	resourceManagerEndpointURL := Environments[environment].ResourceManagerEndpointURL
	spt, err := adal.NewServicePrincipalToken(*oauthConfig, clientID, clientSecret, resourceManagerEndpointURL, nil)
	if err != nil {
		logger.Error("new-oauth-service-principal-token", err, lager.Data{
			"Environment":                environment,
			"resourceManagerEndpointURL": resourceManagerEndpointURL,
			"TenanID":                    tenantID,
			"ClientID":                   clientID,
		})
		return fmt.Errorf("Error in initManagementClient: %v", err)
	}
	client := resources.NewDeploymentsClientWithBaseURI(resourceManagerEndpointURL, account.SubscriptionID)
	account.resourcesManagementClient = &client
	account.resourcesManagementClient.Authorizer = autorest.NewBearerAuthorizer((spt))
	return nil
}

func (account *AzureAccount) Create() (<-chan resources.DeploymentExtended, <-chan error, error) {
	logger := account.logger.Session("Craete")
	logger.Info("Start")
	defer logger.Info("end")

	var parameters map[string]interface{}
	jsonParameters, err := json.Marshal(account.blockchainconfiguration)
	if err != nil {
		logger.Error("convert-config-to-json", err, lager.Data{
			"Configuration": account.blockchainconfiguration,
		})
		return nil, nil, fmt.Errorf("Error in Create: %v", err)
	}
	err = json.Unmarshal(jsonParameters, parameters)
	if err != nil {
		logger.Error("convert-json-to-map[string]interface{}", err, lager.Data{
			"Configuration": account.blockchainconfiguration,
		})
		return nil, nil, fmt.Errorf("Error in Create: %v", err)
	}
	deploymentProps := resources.DeploymentProperties{
		TemplateLink: &resources.TemplateLink{
			URI:            to.StringPtr("https://raw.githubusercontent.com/Azure/azure-quickstart-templates/master/ethereum-consortium-blockchain-network/azuredeploy.json"),
			ContentVersion: to.StringPtr("1.0.0.0"),
		},
		Parameters: &parameters,
		Mode:       resources.Incremental,
	}
	cancel := make(chan struct{})
	deploymentExtended, chanerr := account.resourcesManagementClient.CreateOrUpdate(account.ResourceGroupName,
		deploymentName,
		resources.Deployment{Properties: &deploymentProps},
		cancel)
	return deploymentExtended, chanerr, nil
}

func (account *AzureAccount) Delete() (<-chan autorest.Response, <-chan error) {
	logger := account.logger.Session("Delete")
	logger.Info("start")
	defer logger.Info("end")

	cancel := make(chan struct{})
	return account.resourcesManagementClient.Delete(account.ResourceGroupName, deploymentName, cancel)
}
