package broker

import (
	"errors"
	"strings"
)

type CloudConfig struct {
	Azure      AzureConfig
	Control    ControlConfig
	AzureStack AzureStackConfig
}

type AzureConfig struct {
	Environment              string
	TenanID                  string
	ClientID                 string
	ClientSecret             string
	DefaultSubscriptionID    string
	DefaultResourceGroupName string
}

type Configuration struct {
	SubscriptionID    string `json:"subscription_id"`
	ResourceGroupName string `json:"resource_group_name"`
	UseHTTPS          string `json:"use_https"` // bool
	Location          string `json:"location"`
	CustomDomainName  string `json:"custom_domain_name"`
	UseSubDomain      string `json:"use_sub_domain"`    // bool
	EnableEncryption  string `json:"enable_encryption"` // bool
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

type ControlConfig struct {
	AllowCreateStorageAccount bool
	AllowCreateFileShare      bool
	AllowDeleteStorageAccount bool
	AllowDeleteFileShare      bool
}

func NewControlConfig(allowCreateStorageAccount, allowCreateFileShare, allowDeleteStorageAccount, allowDeleteFileShare bool) *ControlConfig {
	myConf := new(ControlConfig)

	// TBD: Now this broker does not support to create storage account if it does not exist
	myConf.AllowCreateStorageAccount = false
	myConf.AllowCreateFileShare = allowCreateFileShare
	myConf.AllowDeleteStorageAccount = allowDeleteStorageAccount
	myConf.AllowDeleteFileShare = allowDeleteFileShare

	return myConf
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
