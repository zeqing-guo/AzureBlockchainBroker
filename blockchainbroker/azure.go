package blockchainbroker

import (
	"fmt"

	"code.cloudfoundry.org/lager"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
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

type Account struct {
	logger            lager.Logger
	cloudConfig       *CloudConfig
	SubscriptionID    string
	ResourceGroupName string
	UseHTTPS          bool
	// SkuName                  storage.SkuName
	Location                string
	CustomDomainName        string
	UseSubDomain            bool
	EnableEncryption        bool
	IsCreatedStorageAccount bool
	accessKey               string
	baseURL                 string
}

// func NewAccount(logger lager.Logger, cloudConfig *CloudConfig, configuretion Configuration) (*Account, error) {
//
// }
