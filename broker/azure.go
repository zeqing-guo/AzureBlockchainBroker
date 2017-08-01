package broker

import (
	"errors"
	"fmt"
	"time"

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
	groupsClient              *resources.GroupsClient
	DeploymentResult          <-chan resources.DeploymentExtended
	DeploymentError           <-chan error
	blockchainConfiguration   *BlockchainConfiguration
}

func NewAzureAccount(logger lager.Logger, cloudConfig CloudConfig, configuration Configuration, blockchainConfiguration BlockchainConfiguration) (*AzureAccount, error) {
	logger = logger.Session("new-azure-account")
	logger.Info("start")
	defer logger.Info("end", nil)
	azureAccount := AzureAccount{
		logger:                    logger,
		cloudConfig:               &cloudConfig,
		SubscriptionID:            configuration.SubscriptionID,
		ResourceGroupName:         configuration.ResourceGroupName,
		UseHTTPS:                  configuration.UseHTTPS,
		Location:                  locationWestUS,
		CustomDomainName:          "",
		UseSubDomain:              configuration.UseSubDomain,
		EnableEncryption:          configuration.EnableEncryption,
		baseURL:                   "",
		resourcesManagementClient: nil,
		blockchainConfiguration:   &blockchainConfiguration,
		DeploymentResult:          nil,
		DeploymentError:           nil,
	}
	if configuration.Location != "" {
		azureAccount.Location = configuration.Location
	}

	if err := azureAccount.initManagementClient(); err != nil {
		return nil, err
	}
	return &azureAccount, nil
}

func (account *AzureAccount) initManagementClient() error {
	logger := account.logger.Session("init-management-client")
	logger.Info("start")
	defer logger.Info("end")

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
	spt, err := adal.NewServicePrincipalToken(*oauthConfig, clientID, clientSecret, resourceManagerEndpointURL)
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
	account.resourcesManagementClient.Authorizer = autorest.NewBearerAuthorizer(spt)
	groupsClient := resources.NewGroupsClientWithBaseURI(resourceManagerEndpointURL, account.SubscriptionID)
	account.groupsClient = &groupsClient
	account.groupsClient.Authorizer = autorest.NewBearerAuthorizer(spt)
	return nil
}

func (account *AzureAccount) Exist() (bool, error) {
	logger := account.logger.Session("Exist")
	logger.Info("start")
	defer logger.Info("end")

	result, err := account.groupsClient.CheckExistence(account.ResourceGroupName)
	if err != nil {
		logger.Error("error-in-check-exist", err, lager.Data{
			"ResourceGroupName": account.ResourceGroupName,
		})
		return false, fmt.Errorf("Error in broker.Exist: %v", err)
	}
	if result.StatusCode < 400 {
		// 20X for existence
		return true, nil
	}
	// 40X for not existence
	return false, nil
}

func (account *AzureAccount) Create() error {
	logger := account.logger.Session("Craete")
	logger.Info("Start")
	defer logger.Info("end")

	// check if group exists
	parameter := resources.Group{
		Location: &account.Location,
	}

	existed, err := account.Exist()
	if err != nil {
		return err
	}
	if !existed {
		_, err := account.groupsClient.CreateOrUpdate(account.ResourceGroupName, parameter)
		if err != nil {
			logger.Error("error-in-create-resource-group", err, lager.Data{
				"ResourceGroupName": account.ResourceGroupName,
				"parameter":         parameter,
			})
			return fmt.Errorf("Error in broker.Create: %v", err)
		}
	}

	parameters := struct2map(*account.blockchainConfiguration)
	deploymentProps := resources.DeploymentProperties{
		TemplateLink: &resources.TemplateLink{
			URI:            to.StringPtr("https://raw.githubusercontent.com/Azure/azure-quickstart-templates/master/ethereum-consortium-blockchain-network/azuredeploy.json"),
			ContentVersion: to.StringPtr("1.0.0.0"),
		},
		Parameters: &parameters,
		Mode:       resources.Incremental,
	}
	cancel := make(chan struct{})
	account.resourcesManagementClient.CreateOrUpdate(account.ResourceGroupName,
		deploymentName,
		resources.Deployment{Properties: &deploymentProps},
		cancel)
	// err = <-errchan
	// if err != nil {
	// 	return fmt.Errorf("Error in broker.Create: %v", err)
	// }

	cancel <- struct{}{}
	return nil
}

func (account *AzureAccount) Delete() (<-chan autorest.Response, <-chan error) {
	logger := account.logger.Session("Delete")
	logger.Info("start")
	defer logger.Info("end")

	cancel := make(chan struct{})

	return account.resourcesManagementClient.Delete(account.ResourceGroupName, deploymentName, cancel)
}

func (account *AzureAccount) Get() (result resources.DeploymentExtended, err error) {
	return account.resourcesManagementClient.Get(account.ResourceGroupName, deploymentName)
}

func isAccepted(cancel chan struct{}, fn func() (bool, error)) (bool, error) {
	timeout := time.After(5 * time.Second)
	tick := time.Tick(500 * time.Millisecond)

	for {
		select {
		case <-timeout:
			cancel <- struct{}{}
			return false, errors.New("request time out")
		case <-tick:
			ok, err := fn()
			if err != nil {
				return false, err
			}
			if ok {
				return ok, nil
			}
		}
	}
}

func struct2map(blockchainConfiguration BlockchainConfiguration) map[string]interface{} {
	result := make(map[string]interface{})

	result["namePrefix"] = makeStringParameterValueString(blockchainConfiguration.namePrefix)
	result["adminUsername"] = makeStringParameterValueString(blockchainConfiguration.adminUsername)
	result["adminPassword"] = makeStringParameterValueString(blockchainConfiguration.adminPassword)
	result["ethereumAccountPsswd"] = makeStringParameterValueString(blockchainConfiguration.ethereumAccountPsswd)
	result["ethereumAccountPassphrase"] = makeStringParameterValueString(blockchainConfiguration.ethereumAccountPassphrase)
	result["ethereumNetworkID"] = makeStringParameterValueUint64(blockchainConfiguration.ethereumNetworkID)
	result["numConsortiumMembers"] = makeStringParameterValueUint64(blockchainConfiguration.numConsortiumMembers)
	result["numMiningNodesPerMember"] = makeStringParameterValueUint64(blockchainConfiguration.numMiningNodesPerMember)
	result["mnNodeVMSize"] = makeStringParameterValueString(blockchainConfiguration.mnNodeVMSize)
	result["numTXNodes"] = makeStringParameterValueUint64(blockchainConfiguration.numTXNodes)
	result["txNodeVMSize"] = makeStringParameterValueString(blockchainConfiguration.txNodeVMSize)

	return result
}

func makeStringParameterValueString(value string) map[string]string {
	result := make(map[string]string)
	result["value"] = value
	return result
}

func makeStringParameterValueUint64(value uint64) map[string]uint64 {
	result := make(map[string]uint64)
	result["value"] = value
	return result
}
