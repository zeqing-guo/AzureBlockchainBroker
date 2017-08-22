package blockchainbroker

type CloudConfig struct {
	Azure      AzureConfig
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

type AzureStackConfig struct {
	AzureStackDomain         string
	AzureStackAuthentication string
	AzureStackResource       string
	AzureStackEndpointPrefix string
}
