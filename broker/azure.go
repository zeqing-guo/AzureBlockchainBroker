package broker

import (
	// "errors"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	resty "gopkg.in/resty.v0"

	"code.cloudfoundry.org/lager"
	// "github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	// "github.com/Azure/go-autorest/autorest"
	// "github.com/Azure/go-autorest/autorest/adal"
	// "github.com/Azure/go-autorest/autorest/to"
)

const (
	creator                     = "creator"
	resourceNotFound            = "StatusCode=404"
	fileRequestTimeoutInSeconds = 60
	locationWestUS              = "westus"
	blockchainTemplate          = "https://github.com/Azure/azure-quickstart-templates/raw/master/ethereum-consortium-blockchain-network/azuredeploy.json"
	templateVersion             = "1.0.0.0"
)

var (
	restRetryCodes = []int{408, 429, 500, 502, 503, 504}
)

var (
	userAgent          = "azureblockchainbroker"
	contentTypeJSON    = "application/json"
	contentTypeWWW     = "application/x-www-form-urlencoded"
	restAPIProvider    = "Microsoft.Resources"
	restAPIDeployments = "deployments"
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
	Template        string
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
			Template:        "2017-05-10",
			Storage:         "2016-05-31",
			Group:           "2017-05-10",
			ActiveDirectory: "2015-06-15",
		},
	},
	AzureChinaCloud: Environment{
		ResourceManagerEndpointURL: "https://management.chinacloudapi.cn/",
		ActiveDirectoryEndpointURL: "https://login.chinacloudapi.cn",
		APIVersions: APIVersions{
			Template:        "2017-05-10",
			Storage:         "2016-05-31",
			Group:           "2017-05-10",
			ActiveDirectory: "2015-06-15",
		},
	},
	AzureUSGovernment: Environment{
		ResourceManagerEndpointURL: "https://management.usgovcloudapi.net/",
		ActiveDirectoryEndpointURL: "https://login.microsoftonline.com",
		APIVersions: APIVersions{
			Template:        "2017-05-10",
			Storage:         "2016-05-31",
			Group:           "2017-05-10",
			ActiveDirectory: "2015-06-15",
		},
	},
	AzureGermanCloud: Environment{
		ResourceManagerEndpointURL: "https://management.microsoftazure.de/",
		ActiveDirectoryEndpointURL: "https://login.microsoftonline.de",
		APIVersions: APIVersions{
			Template:        "2017-05-10",
			Storage:         "2015-06-15",
			Group:           "2017-05-10",
			ActiveDirectory: "2015-06-15",
		},
	},
	AzureStack: Environment{
		APIVersions: APIVersions{
			Template:        "2017-05-10",
			Storage:         "2015-06-15",
			Group:           "2017-05-10",
			ActiveDirectory: "2015-06-15",
		},
	},
}

type AzureToken struct {
	ExpiresOn   time.Time
	AccessToken string
}

type AzureRESTClient struct {
	logger         lager.Logger
	cloudConfig    *CloudConfig
	resourceConfig *ResourceConfig
	token          AzureToken
}

func NewAzureResourceAccountRESTClient(logger lager.Logger, cloudConfig *CloudConfig, resourceConfig *ResourceConfig) (AzureRESTClient, error) {
	logger = logger.Session("create-resource-account-rest-client")
	client := AzureRESTClient{
		logger:         logger,
		cloudConfig:    cloudConfig,
		resourceConfig: resourceConfig,
	}
	return client, nil
}

func (c *AzureRESTClient) refreshToken(force bool) error {
	if c.token.AccessToken == "" || time.Until(c.token.ExpiresOn) <= 0 || force {
		// fmt.Println("get token")
		headers := map[string]string{
			"Content-Type": contentTypeWWW,
			"User-Agent":   userAgent,
		}

		hostURL := fmt.Sprintf("%s/%s/oauth2/token", Environments[c.cloudConfig.Azure.Environment].ActiveDirectoryEndpointURL, c.cloudConfig.Azure.TenanID)
		body := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {c.cloudConfig.Azure.ClientID},
			"client_secret": {c.cloudConfig.Azure.ClientSecret},
			"resource":      {Environments[c.cloudConfig.Azure.Environment].ResourceManagerEndpointURL},
			"scope":         {"user_impersonation"},
		}

		resty.DefaultClient.SetRetryCount(3).SetRetryWaitTime(10)
		resp, err := resty.R().
			SetHeaders(headers).
			SetQueryParam("api-version", Environments[c.cloudConfig.Azure.Environment].APIVersions.ActiveDirectory).
			SetBody(body.Encode()).
			Post(hostURL)
		if err != nil {
			return err
		}
		if resp.StatusCode() == http.StatusOK {
			type ResponseBody struct {
				ExpiresOn   string `json:"expires_on"`
				AccessToken string `json:"access_token"`
			}
			responseBody := ResponseBody{}
			err := json.Unmarshal(resp.Body(), &responseBody)
			if err != nil {
				return err
			}
			expiresOn, err := strconv.ParseInt(responseBody.ExpiresOn, 10, 64)
			if err != nil {
				return err
			}
			c.token.ExpiresOn = time.Unix(expiresOn, 0)
			c.token.AccessToken = responseBody.AccessToken
		} else {
			return fmt.Errorf("HTTP CODE: %#v", resp.StatusCode())
		}
	}
	return nil
}

func (c *AzureRESTClient) initialize() (headers map[string]string, err error) {
	resty.DefaultClient.SetRetryCount(3).SetRetryWaitTime(10)
	check := resty.RetryConditionFunc(func(r *resty.Response) (bool, error) {
		for _, v := range restRetryCodes {
			if r.StatusCode() == v {
				return true, nil
			}
		}
		return false, nil
	})
	resty.DefaultClient.AddRetryCondition(check)
	headers = map[string]string{
		"Content-Type": contentTypeJSON,
		"User-Agent":   userAgent,
	}
	err = c.refreshToken(false)
	if err != nil {
		return nil, err
	}

	return headers, nil
}

func (c *AzureRESTClient) GroupExist() (bool, error) {
	headers, err := c.initialize()
	if err != nil {
		return false, err
	}
	queries := map[string]string{
		"api-version": Environments[c.cloudConfig.Azure.Environment].APIVersions.Group,
	}
	hostURL := fmt.Sprintf(
		"%s/subscriptions/%s/resourceGroups/%s",
		Environments[c.cloudConfig.Azure.Environment].ResourceManagerEndpointURL,
		c.resourceConfig.SubscriptionID,
		c.resourceConfig.ResourceGroupName,
	)
	resp, err := resty.R().
		SetHeaders(headers).
		SetQueryParams(queries).
		SetAuthToken(c.token.AccessToken).
		Head(hostURL)
	if err != nil {
		return false, err
	}
	// 204 for existing, 404 for not found
	return resp.StatusCode() == http.StatusNoContent, nil
}

func (c *AzureRESTClient) CreateGroup() (bool, error) {
	headers, err := c.initialize()
	if err != nil {
		return false, err
	}
	queries := map[string]string{
		"api-version": Environments[c.cloudConfig.Azure.Environment].APIVersions.Group,
	}

	hostURL := fmt.Sprintf(
		"%s/subscriptions/%s/resourcegroups/%s",
		Environments[c.cloudConfig.Azure.Environment].ResourceManagerEndpointURL,
		c.resourceConfig.SubscriptionID,
		c.resourceConfig.ResourceGroupName,
	)
	resourceGroup := map[string]interface{}{
		"location": c.resourceConfig.Location,
	}
	body, err := json.Marshal(resourceGroup)

	resp, err := resty.R().
		SetHeaders(headers).
		SetQueryParams(queries).
		SetAuthToken(c.token.AccessToken).
		SetBody(body).
		Put(hostURL)
	if err != nil {
		return false, err
	}
	statusCode := resp.StatusCode()
	if statusCode == http.StatusOK || statusCode == http.StatusCreated {
		return true, nil
	}
	return false, fmt.Errorf("Error Code: %d, %v", statusCode, resp)
}

func (c *AzureRESTClient) CheckResourceStatus(resourceGroupName string) (string, error) {
	headers, err := c.initialize()
	if err != nil {
		return "", err
	}
	queries := map[string]string{
		"api-version": Environments[c.cloudConfig.Azure.Environment].APIVersions.Group,
	}
	hostURL := fmt.Sprintf("%s/subscriptions/%s/resourceGroups/%s",
		Environments[c.cloudConfig.Azure.Environment].ResourceManagerEndpointURL,
		c.resourceConfig.SubscriptionID,
		resourceGroupName,
	)

	resp, err := resty.R().
		SetHeaders(headers).
		SetQueryParams(queries).
		SetAuthToken(c.token.AccessToken).
		Get(hostURL)

	statusCode := resp.StatusCode()
	if statusCode == http.StatusOK {
		apiResponse := map[string]interface{}{}
		if err := json.Unmarshal(resp.Body(), &apiResponse); err != nil {
			return "", fmt.Errorf("StatusCode: %d - %v\n\t%s", statusCode, resp, err)
		}
		properties := apiResponse["properties"].(map[string]interface{})
		return properties["provisioningState"].(string), nil
	} else if statusCode == http.StatusNotFound {
		return "notfound", nil
	}
	return "", fmt.Errorf("StatusCode: %d - %v", statusCode, resp)
}

func (c *AzureRESTClient) DeleteGroup() (bool, error) {
	headers, err := c.initialize()
	if err != nil {
		return false, err
	}
	queries := map[string]string{
		"api-version": Environments[c.cloudConfig.Azure.Environment].APIVersions.Group,
	}
	hostURL := fmt.Sprintf(
		"%s/subscriptions/%s/resourcegroups/%s",
		Environments[c.cloudConfig.Azure.Environment].ResourceManagerEndpointURL,
		c.resourceConfig.SubscriptionID,
		c.resourceConfig.ResourceGroupName,
	)

	resp, err := resty.R().
		SetHeaders(headers).
		SetQueryParams(queries).
		SetAuthToken(c.token.AccessToken).
		Delete(hostURL)
	if err != nil {
		return false, err
	}
	statusCode := resp.StatusCode()
	if statusCode == http.StatusOK || statusCode == http.StatusAccepted {
		return true, nil
	}
	return false, fmt.Errorf("Error Code: %d, %v", statusCode, resp)
}

// resource management: deployments
func (c *AzureRESTClient) DeleteResource(deploymentName string) (bool, error) {
	headers, err := c.initialize()
	if err != nil {
		return false, err
	}
	queries := map[string]string{
		"api-version": Environments[c.cloudConfig.Azure.Environment].APIVersions.Group,
	}
	hostURL := fmt.Sprintf("%s/subscriptions/%s/resourceGroups/%s/providers/%s/%s/%s",
		Environments[c.cloudConfig.Azure.Environment].ResourceManagerEndpointURL,
		c.resourceConfig.SubscriptionID,
		c.resourceConfig.ResourceGroupName,
		restAPIProvider,
		restAPIDeployments,
		deploymentName,
	)

	resp, err := resty.R().
		SetHeaders(headers).
		SetQueryParams(queries).
		SetAuthToken(c.token.AccessToken).
		Delete(hostURL)
	statusCode := resp.StatusCode()
	if statusCode == http.StatusAccepted {
		return true, nil
	} else if statusCode == http.StatusNoContent {
		return true, nil
	}
	return false, fmt.Errorf("Error Code: %d, %v", statusCode, resp)
}

func (c *AzureRESTClient) DeployTemplate(deploymentName string, template *map[string]interface{}, templateLink *Link, parameters *map[string]interface{}, parametersLink *Link) (string, error) {
	mode := "Incremental"
	headers, err := c.initialize()
	queries := map[string]string{
		"api-version": Environments[c.cloudConfig.Azure.Environment].APIVersions.Template,
	}
	if err != nil {
		return "", err
	}
	hostURL := fmt.Sprintf("%s/subscriptions/%s/resourceGroups/%s/providers/%s/%s/%s",
		Environments[c.cloudConfig.Azure.Environment].ResourceManagerEndpointURL,
		c.resourceConfig.SubscriptionID,
		c.resourceConfig.ResourceGroupName,
		restAPIProvider,
		restAPIDeployments,
		deploymentName,
	)
	tags := map[string]string{}
	tags["User-Agent"] = userAgent

	var properties map[string]interface{}
	if template != nil {
		if parameters != nil {
			properties = map[string]interface{}{
				"template":   template,
				"parameters": parameters,
				"mode":       mode,
			}
		} else {
			properties = map[string]interface{}{
				"template": template,
				"parametersLink": map[string]interface{}{
					"uri":            parametersLink.uri,
					"contentVersion": parametersLink.contentVersion,
				},
				"mode": mode,
			}
		}
	} else {
		if parameters != nil {
			properties = map[string]interface{}{
				"templateLink": map[string]interface{}{
					"uri":            templateLink.uri,
					"contentVersion": templateLink.contentVersion,
				},
				"parameters": parameters,
				"mode":       mode,
			}
		} else {
			properties = map[string]interface{}{
				"templateLink": map[string]interface{}{
					"uri":            templateLink.uri,
					"contentVersion": templateLink.contentVersion,
				},
				"parametersLink": map[string]interface{}{
					"uri":            parametersLink.uri,
					"contentVersion": parametersLink.contentVersion,
				},
				"mode": mode,
			}

		}
	}
	deployTemplate := map[string]interface{}{
		"properties": properties,
	}
	body, err := json.Marshal(deployTemplate)
	if err != nil {
		return "", err
	}

	resp, err := resty.R().
		SetHeaders(headers).
		SetQueryParams(queries).
		SetAuthToken(c.token.AccessToken).
		SetBody(body).
		Put(hostURL)
	if err != nil {
		return "", err
	}
	statusCode := resp.StatusCode()
	// the doc regards 200 and 201 as the same thing
	// https://docs.microsoft.com/en-us/rest/api/resources/deployments/createorupdate#deployments_createorupdate_responses
	if statusCode == http.StatusOK {
		return "ok", nil
	} else if statusCode == http.StatusCreated {
		return "created", nil
	} else {
		return "", fmt.Errorf("Error Code: %d, %v", statusCode, resp)
	}
}

func (c *AzureRESTClient) GetStatusURL(deploymentName string) (string, error) {
	hostURL := fmt.Sprintf("%s/subscriptions/%s/resourceGroups/%s/providers/%s/%s/%s?api-version=%s",
		Environments[c.cloudConfig.Azure.Environment].ResourceManagerEndpointURL,
		c.resourceConfig.SubscriptionID,
		c.resourceConfig.ResourceGroupName,
		restAPIProvider,
		restAPIDeployments,
		deploymentName,
		Environments[c.cloudConfig.Azure.Environment].APIVersions.Group,
	)
	return hostURL, nil
}

func (c *AzureRESTClient) GetAdminAndRPCUrl(deploymentName string) (adminSiteURL string, rpcURL string, err error) {
	logger := c.logger
	logger.Info("start")
	defer logger.Info("end")
	hostURL, _ := c.GetStatusURL(deploymentName)
	resp, err := c.get(hostURL)

	statusCode := resp.StatusCode()
	if statusCode != http.StatusOK {
		return "", "", fmt.Errorf("Error Code: %d, %v", statusCode, resp)
	}
	apiResponse := map[string]interface{}{}
	if err := json.Unmarshal(resp.Body(), &apiResponse); err != nil {
		return "", "", fmt.Errorf("StatusCode: %d - %v\n\t%s", statusCode, resp, err)
	}
	properties := apiResponse["properties"].(map[string]interface{})
	outputs := properties["outputs"].(map[string]interface{})
	adminSite := outputs["admin-site"].(map[string]interface{})
	ethereumRPCEndpoint := outputs["ethereum-rpc-endpoint"].(map[string]interface{})
	return adminSite["value"].(string), ethereumRPCEndpoint["value"].(string), nil
}

func (c *AzureRESTClient) CheckCompletion(deploymentName string) (string, error) {
	logger := c.logger
	logger.Info("start")
	defer logger.Info("end")
	hostURL, _ := c.GetStatusURL(deploymentName)
	resp, err := c.get(hostURL)
	if err != nil {
		return "", err
	}

	statusCode := resp.StatusCode()
	if statusCode != http.StatusOK {
		return "", fmt.Errorf("Error Code: %d, %v", statusCode, resp)
	}
	apiResponse := map[string]interface{}{}
	if err := json.Unmarshal(resp.Body(), &apiResponse); err != nil {
		return "", fmt.Errorf("StatusCode: %d - %v\n\t%s", statusCode, resp, err)
	}
	provisioningState := apiResponse["properties"].(map[string]interface{})["provisioningState"]
	return provisioningState.(string), nil
}

func (c *AzureRESTClient) get(asyncURL string) (*resty.Response, error) {
	headers, err := c.initialize()
	if err != nil {
		return nil, err
	}
	queries := map[string]string{
		"api-version": Environments[c.cloudConfig.Azure.Environment].APIVersions.Group,
	}

	return resty.R().
		SetHeaders(headers).
		SetQueryParams(queries).
		SetAuthToken(c.token.AccessToken).
		Get(asyncURL)
}

type DeploymentClient struct {
	logger           lager.Logger
	blockchainConfig BlockchainConfig
	azureRESTClient  *AzureRESTClient
}

func NewDeploymentClient(logger lager.Logger, cloudConfig CloudConfig, resourceConfig ResourceConfig, blockchainConfig BlockchainConfig) (*DeploymentClient, error) {
	logger = logger.Session("new-deployment-client")
	logger.Info("start")
	defer logger.Info("end")
	deploymentClient := DeploymentClient{
		logger:           logger,
		blockchainConfig: blockchainConfig,
		azureRESTClient:  nil,
	}
	err := deploymentClient.initialize(cloudConfig, resourceConfig)
	if err != nil {
		logger.Error("error-when-initialize-deploy-client", err)
		return nil, err
	}
	return &deploymentClient, nil
}

func (d *DeploymentClient) initialize(cloudConfig CloudConfig, resourceConfig ResourceConfig) (err error) {
	logger := d.logger.Session("init-deployment-client")
	logger.Info("start")
	defer logger.Info("end")

	// create REST client for async operation
	azureRESTClient, err := NewAzureResourceAccountRESTClient(d.logger, &cloudConfig, &resourceConfig)
	if err != nil {
		return fmt.Errorf("Error in initialize DeploymentClient: %v", err)
	}
	d.azureRESTClient = &azureRESTClient
	return nil
}

func (d *DeploymentClient) Create(deploymentName string) (err error) {
	logger := d.logger.Session("create-template")
	logger.Info("start")
	defer logger.Info("end")

	// check resource group whether existing, if not create it
	azureRESTClient := d.azureRESTClient
	exist, err := azureRESTClient.GroupExist()
	if err != nil {
		return fmt.Errorf("Error in check GroupExist: %v", err)
	}
	if !exist {
		_, err := azureRESTClient.CreateGroup()
		if err != nil {
			return fmt.Errorf("Error in create group: %v", err)
		}
	}

	// deploy template
	parameters := struct2map(d.blockchainConfig)
	link := NewLink(blockchainTemplate, templateVersion)
	_, err = azureRESTClient.DeployTemplate(deploymentName, nil, link, &parameters, nil)
	if err != nil {
		return fmt.Errorf("Error in deploy template: %v", err)
	}
	return nil
}

func struct2map(blockchainConfig BlockchainConfig) map[string]interface{} {
	result := make(map[string]interface{})

	result["namePrefix"] = makeStringParameterValueString(blockchainConfig.namePrefix)
	result["adminUsername"] = makeStringParameterValueString(blockchainConfig.adminUsername)
	result["adminPassword"] = makeStringParameterValueString(blockchainConfig.adminPassword)
	result["ethereumAccountPsswd"] = makeStringParameterValueString(blockchainConfig.ethereumAccountPsswd)
	result["ethereumAccountPassphrase"] = makeStringParameterValueString(blockchainConfig.ethereumAccountPassphrase)
	result["ethereumNetworkID"] = makeStringParameterValueUint64(blockchainConfig.ethereumNetworkID)
	result["numConsortiumMembers"] = makeStringParameterValueUint64(blockchainConfig.numConsortiumMembers)
	result["numMiningNodesPerMember"] = makeStringParameterValueUint64(blockchainConfig.numMiningNodesPerMember)
	result["mnNodeVMSize"] = makeStringParameterValueString(blockchainConfig.mnNodeVMSize)
	result["numTXNodes"] = makeStringParameterValueUint64(blockchainConfig.numTXNodes)
	result["txNodeVMSize"] = makeStringParameterValueString(blockchainConfig.txNodeVMSize)

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

// type AzureAccount struct {
// 	logger                    lager.Logger
// 	cloudConfig               *CloudConfig
// 	SubscriptionID            string
// 	ResourceGroupName         string
// 	UseHTTPS                  bool
// 	Location                  string
// 	CustomDomainName          string
// 	UseSubDomain              bool
// 	EnableEncryption          bool
// 	baseURL                   string
// 	resourcesManagementClient *resources.DeploymentsClient
// 	groupsClient              *resources.GroupsClient
// 	DeploymentResult          <-chan resources.DeploymentExtended
// 	DeploymentError           <-chan error
// 	blockchainConfig   *BlockchainConfiguration
// }
//
// func NewAzureAccount(logger lager.Logger, cloudConfig CloudConfig, configuration Configuration, blockchainConfig BlockchainConfiguration) (*AzureAccount, error) {
// 	logger = logger.Session("new-azure-account")
// 	logger.Info("start")
// 	defer logger.Info("end", nil)
// 	azureAccount := AzureAccount{
// 		logger:                    logger,
// 		cloudConfig:               &cloudConfig,
// 		SubscriptionID:            configuration.SubscriptionID,
// 		ResourceGroupName:         configuration.ResourceGroupName,
// 		UseHTTPS:                  configuration.UseHTTPS,
// 		Location:                  locationWestUS,
// 		CustomDomainName:          "",
// 		UseSubDomain:              configuration.UseSubDomain,
// 		EnableEncryption:          configuration.EnableEncryption,
// 		baseURL:                   "",
// 		resourcesManagementClient: nil,
// 		blockchainConfig:   &blockchainConfig,
// 		DeploymentResult:          nil,
// 		DeploymentError:           nil,
// 	}
// 	if configuration.Location != "" {
// 		azureAccount.Location = configuration.Location
// 	}
//
// 	if err := azureAccount.initManagementClient(); err != nil {
// 		return nil, err
// 	}
// 	return &azureAccount, nil
// }
//
// func (account *AzureAccount) initManagementClient() error {
// 	logger := account.logger.Session("init-management-client")
// 	logger.Info("start")
// 	defer logger.Info("end")
//
// 	environment := account.cloudConfig.Azure.Environment
// 	tenantID := account.cloudConfig.Azure.TenanID
// 	clientID := account.cloudConfig.Azure.ClientID
// 	clientSecret := account.cloudConfig.Azure.ClientSecret
// 	oauthConfig, err := adal.NewOAuthConfig(Environments[environment].ActiveDirectoryEndpointURL, tenantID)
// 	if err != nil {
// 		logger.Error("new-oauth-config", err, lager.Data{
// 			"Environment":                environment,
// 			"ActiveDirectoryEndpointURL": Environments[environment].ActiveDirectoryEndpointURL,
// 			"TenanID":                    tenantID,
// 		})
// 		return fmt.Errorf("Error in initManagementClient: %v", err)
// 	}
//
// 	resourceManagerEndpointURL := Environments[environment].ResourceManagerEndpointURL
// 	spt, err := adal.NewServicePrincipalToken(*oauthConfig, clientID, clientSecret, resourceManagerEndpointURL)
// 	if err != nil {
// 		logger.Error("new-oauth-service-principal-token", err, lager.Data{
// 			"Environment":                environment,
// 			"resourceManagerEndpointURL": resourceManagerEndpointURL,
// 			"TenanID":                    tenantID,
// 			"ClientID":                   clientID,
// 		})
// 		return fmt.Errorf("Error in initManagementClient: %v", err)
// 	}
// 	client := resources.NewDeploymentsClientWithBaseURI(resourceManagerEndpointURL, account.SubscriptionID)
// 	account.resourcesManagementClient = &client
// 	account.resourcesManagementClient.Authorizer = autorest.NewBearerAuthorizer(spt)
// 	groupsClient := resources.NewGroupsClientWithBaseURI(resourceManagerEndpointURL, account.SubscriptionID)
// 	account.groupsClient = &groupsClient
// 	account.groupsClient.Authorizer = autorest.NewBearerAuthorizer(spt)
// 	return nil
// }
//
// func (account *AzureAccount) Exist() (bool, error) {
// 	logger := account.logger.Session("Exist")
// 	logger.Info("start")
// 	defer logger.Info("end")
//
// 	result, err := account.groupsClient.CheckExistence(account.ResourceGroupName)
// 	if err != nil {
// 		logger.Error("error-in-check-exist", err, lager.Data{
// 			"ResourceGroupName": account.ResourceGroupName,
// 		})
// 		return false, fmt.Errorf("Error in broker.Exist: %v", err)
// 	}
// 	if result.StatusCode < 400 {
// 		// 20X for existence
// 		return true, nil
// 	}
// 	// 40X for not existence
// 	return false, nil
// }
//
// func (account *AzureAccount) Create() error {
// 	logger := account.logger.Session("Craete")
// 	logger.Info("Start")
// 	defer logger.Info("end")
//
// 	// check if group exists
// 	parameter := resources.Group{
// 		Location: &account.Location,
// 	}
//
// 	existed, err := account.Exist()
// 	if err != nil {
// 		return err
// 	}
// 	if !existed {
// 		_, err := account.groupsClient.CreateOrUpdate(account.ResourceGroupName, parameter)
// 		if err != nil {
// 			logger.Error("error-in-create-resource-group", err, lager.Data{
// 				"ResourceGroupName": account.ResourceGroupName,
// 				"parameter":         parameter,
// 			})
// 			return fmt.Errorf("Error in broker.Create: %v", err)
// 		}
// 	}
//
// 	parameters := struct2map(*account.blockchainConfig)
// 	deploymentProps := resources.DeploymentProperties{
// 		TemplateLink: &resources.TemplateLink{
// 			URI:            to.StringPtr("https://raw.githubusercontent.com/Azure/azure-quickstart-templates/master/ethereum-consortium-blockchain-network/azuredeploy.json"),
// 			ContentVersion: to.StringPtr("1.0.0.0"),
// 		},
// 		Parameters: &parameters,
// 		Mode:       resources.Incremental,
// 	}
// 	cancel := make(chan struct{})
// 	account.resourcesManagementClient.CreateOrUpdate(account.ResourceGroupName,
// 		deploymentName,
// 		resources.Deployment{Properties: &deploymentProps},
// 		cancel)
// 	// err = <-errchan
// 	// if err != nil {
// 	// 	return fmt.Errorf("Error in broker.Create: %v", err)
// 	// }
//
// 	cancel <- struct{}{}
// 	return nil
// }
//
// func (account *AzureAccount) Delete() (<-chan autorest.Response, <-chan error) {
// 	logger := account.logger.Session("Delete")
// 	logger.Info("start")
// 	defer logger.Info("end")
//
// 	cancel := make(chan struct{})
//
// 	return account.resourcesManagementClient.Delete(account.ResourceGroupName, deploymentName, cancel)
// }
//
// func (account *AzureAccount) Get() (result resources.DeploymentExtended, err error) {
// 	return account.resourcesManagementClient.Get(account.ResourceGroupName, deploymentName)
// }
//
// func isAccepted(cancel chan struct{}, fn func() (bool, error)) (bool, error) {
// 	timeout := time.After(5 * time.Second)
// 	tick := time.Tick(500 * time.Millisecond)
//
// 	for {
// 		select {
// 		case <-timeout:
// 			cancel <- struct{}{}
// 			return false, errors.New("request time out")
// 		case <-tick:
// 			ok, err := fn()
// 			if err != nil {
// 				return false, err
// 			}
// 			if ok {
// 				return ok, nil
// 			}
// 		}
// 	}
// }
