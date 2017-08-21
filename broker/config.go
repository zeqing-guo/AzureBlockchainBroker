package broker

import (
	"errors"
	"strings"
)

type BlockchainConfig struct {
	namePrefix                string
	adminUsername             string
	adminPassword             string
	ethereumAccountPsswd      string
	ethereumAccountPassphrase string
	ethereumNetworkID         uint64
	numConsortiumMembers      uint64
	numMiningNodesPerMember   uint64
	mnNodeVMSize              string
	numTXNodes                uint64
	txNodeVMSize              string
}

func NewBlockchainConfig(
	namePrefix string,
	adminUsername string,
	adminPassword string,
	ethereumAccountPsswd string,
	ethereumAccountPassphrase string,
	ethereumNetworkID uint64,
	numConsortiumMembers uint64,
	numMiningNodesPerMember uint64,
	mnNodeVMSize string,
	numTXNodes uint64,
	txNodeVMSize string,
) *BlockchainConfig {
	blockchainConfig := new(BlockchainConfig)

	blockchainConfig.namePrefix = namePrefix
	blockchainConfig.adminUsername = adminUsername
	blockchainConfig.adminPassword = adminPassword
	blockchainConfig.ethereumAccountPsswd = ethereumAccountPsswd
	blockchainConfig.ethereumAccountPassphrase = ethereumAccountPassphrase
	blockchainConfig.ethereumNetworkID = ethereumNetworkID
	blockchainConfig.numConsortiumMembers = numConsortiumMembers
	blockchainConfig.numMiningNodesPerMember = numMiningNodesPerMember
	blockchainConfig.mnNodeVMSize = mnNodeVMSize
	blockchainConfig.numTXNodes = numTXNodes
	blockchainConfig.txNodeVMSize = txNodeVMSize
	return blockchainConfig
}

type CloudConfig struct {
	Azure      AzureConfig
	AzureStack AzureStackConfig
}

func NewCloudConfig(azureConfig AzureConfig, azureStack AzureStackConfig) *CloudConfig {
	cloudConfig := new(CloudConfig)

	cloudConfig.AzureStack = azureStack
	cloudConfig.Azure = azureConfig
	return cloudConfig
}

type AzureConfig struct {
	Environment              string
	TenanID                  string
	ClientID                 string
	ClientSecret             string
	DefaultSubscriptionID    string
	DefaultResourceGroupName string
}

type ResourceConfig struct {
	SubscriptionID    string `json:"subscription_id"`
	ResourceGroupName string `json:"resource_group_name"`
	UseHTTPS          bool   `json:"use_https"` // bool
	Location          string `json:"location"`
	CustomDomainName  string `json:"custom_domain_name"`
	UseSubDomain      bool   `json:"use_sub_domain"`    // bool
	EnableEncryption  bool   `json:"enable_encryption"` // bool
}

func NewResourceConfig(subscriptionID string, resourceGroupName string, useHTTPS bool, location string, customDomainName string, useSubDomain bool, enableEncryption bool) *ResourceConfig {
	config := new(ResourceConfig)

	config.SubscriptionID = subscriptionID
	config.ResourceGroupName = resourceGroupName
	config.UseHTTPS = useHTTPS
	config.Location = location
	config.CustomDomainName = customDomainName
	config.UseSubDomain = useSubDomain
	config.EnableEncryption = enableEncryption

	return config
}

func NewAzureConfig(environment, tenanID, clientID, clientSecret, defaultSubscriptionID, defaultResourceGroupName string) *AzureConfig {
	myConf := new(AzureConfig)

	myConf.Environment = environment
	myConf.TenanID = tenanID
	myConf.ClientID = clientID
	myConf.ClientSecret = clientSecret
	myConf.DefaultSubscriptionID = defaultSubscriptionID
	myConf.DefaultResourceGroupName = defaultResourceGroupName

	return myConf
}

func (config *AzureConfig) Validate() error {
	missingKeys := []string{}
	if config.Environment == "" {
		missingKeys = append(missingKeys, "environment")
	}
	if config.TenanID == "" {
		missingKeys = append(missingKeys, "tenanID")
	}
	if config.ClientID == "" {
		missingKeys = append(missingKeys, "clientID")
	}
	if config.ClientSecret == "" {
		missingKeys = append(missingKeys, "clientSecret")
	}

	if len(missingKeys) > 0 {
		return errors.New("Missing required parameters: " + strings.Join(missingKeys, ", "))
	}
	return nil
}

type AzureStackConfig struct {
	AzureStackDomain         string
	AzureStackAuthentication string
	AzureStackResource       string
	AzureStackEndpointPrefix string
}

func NewAzureStackConfig(azureStackDomain, azureStackAuthentication, azureStackResource, azureStackEndpointPrefix string) *AzureStackConfig {
	myConf := new(AzureStackConfig)

	myConf.AzureStackDomain = azureStackDomain
	myConf.AzureStackAuthentication = azureStackAuthentication
	myConf.AzureStackResource = azureStackResource
	myConf.AzureStackEndpointPrefix = azureStackEndpointPrefix

	return myConf
}

func (config *AzureStackConfig) Validate() error {
	missingKeys := []string{}
	if config.AzureStackDomain == "" {
		missingKeys = append(missingKeys, "azureStackDomain")
	}
	if config.AzureStackAuthentication == "" {
		missingKeys = append(missingKeys, "azureStackAuthentication")
	}
	if config.AzureStackResource == "" {
		missingKeys = append(missingKeys, "azureStackResource")
	}
	if config.AzureStackEndpointPrefix == "" {
		missingKeys = append(missingKeys, "azureStackEndpointPrefix")
	}

	if len(missingKeys) > 0 {
		return errors.New("Missing required parameters when 'environment' is 'AzureStack': " + strings.Join(missingKeys, ", "))
	}
	return nil
}

type Link struct {
	uri            string
	contentVersion string
}

func NewLink(uri, contentVersion string) (link *Link) {
	link = new(Link)
	link.contentVersion = contentVersion
	link.uri = uri
	return link
}
