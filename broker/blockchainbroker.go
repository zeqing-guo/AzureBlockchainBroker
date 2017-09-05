package broker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-cf/brokerapi"
)

type lock interface {
	Lock()
	Unlock()
}

type ServiceBroker struct {
	logger lager.Logger
	client *DeploymentClient
	static staticState
	mutex  lock
}

type staticState struct {
	ServiceName string `json:"service_name"`
	ServiceID   string `json:"service_id"`
}

func New(logger lager.Logger,
	cloudConfig CloudConfig,
	resourceConfig ResourceConfig,
	blockchainConfig BlockchainConfig,
	serviceName string,
	serviceID string) (*ServiceBroker, error) {
	logger = logger.Session("new-blockchain-service-broker")
	logger.Info("start")
	defer logger.Info("end", nil)
	client, err := NewDeploymentClient(logger, cloudConfig, resourceConfig, blockchainConfig)
	if err != nil {
		return nil, err
	}
	serviceBroker := ServiceBroker{
		logger: logger,
		mutex:  &sync.Mutex{},
		client: client,
		static: staticState{
			ServiceID:   serviceID,
			ServiceName: serviceName,
		},
	}
	return &serviceBroker, nil
}

func (b *ServiceBroker) Services(_ context.Context) []brokerapi.Service {
	logger := b.logger.Session("services")
	logger.Info("start")
	defer logger.Info("end")

	return []brokerapi.Service{{
		ID:            b.static.ServiceID,
		Name:          b.static.ServiceName,
		Description:   "Azure Blockchain",
		Bindable:      true,
		PlanUpdatable: false,
		Tags:          []string{"azureblockchain"},
		Requires:      []brokerapi.RequiredPermission{},
		Plans: []brokerapi.ServicePlan{
			{
				Name:        "AzureBlockchain",
				ID:          "7c0b2254-7e68-11e7-bbe1-000d3a818256",
				Description: "Azure Blockchain",
			},
		},
	}}
}

func (b *ServiceBroker) LastOperation(_ context.Context, instanceID string, operationData string) (brokerapi.LastOperation, error) {
	logger := b.logger.Session("last-operation", lager.Data{"instanceID": instanceID, "operationData": operationData})
	logger.Info("start")
	defer logger.Info("end")

	b.mutex.Lock()
	defer b.mutex.Unlock()

	if operationData == "" {
		return brokerapi.LastOperation{}, errors.New("unrecognized operationData")
	}

	client := *(b.client.azureRESTClient)
	client.resourceConfig.ResourceGroupName = instanceID
	operationDataArr := strings.Split(operationData, ":")
	if len(operationDataArr) != 2 {
		return brokerapi.LastOperation{}, errors.New("unrecognized operationData")
	}
	var state string
	var err error
	if operationDataArr[0] == "provision" {
		state, err = client.CheckCompletion(operationDataArr[1])
	} else if operationDataArr[0] == "deprovision" {
		state, err = client.CheckResourceStatus(operationDataArr[1])
	}

	state = strings.ToLower(state)
	logger.Info("check-state", lager.Data{
		"state": state,
	})
	description := ""
	if err != nil {
		return brokerapi.LastOperation{State: brokerapi.Failed, Description: err.Error()}, nil
	}
	if state == "succeeded" {
		// only provision can return succeeded
		adminSiteURL, rpcURL, err := client.GetAdminAndRPCUrl(operationDataArr[1])
		if err != nil {
			return brokerapi.LastOperation{State: brokerapi.Failed, Description: err.Error()}, nil
		}
		description = fmt.Sprintf("{\"adminSiteURL\": \"%s\", \"rpcURL\": \"%s\"}", adminSiteURL, rpcURL)
		return brokerapi.LastOperation{State: brokerapi.Succeeded, Description: description}, nil
	} else if state == "notfound" && operationDataArr[0] == "deprovision" {
		return brokerapi.LastOperation{State: brokerapi.Succeeded, Description: description}, nil
	} else if state == "failed" {
		return brokerapi.LastOperation{State: brokerapi.Failed, Description: description}, nil
	}

	return brokerapi.LastOperation{State: brokerapi.InProgress, Description: description}, nil
}

func (b *ServiceBroker) Provision(context context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (_ brokerapi.ProvisionedServiceSpec, e error) {
	logger := b.logger.Session("provision").WithData(lager.Data{"instanceID": instanceID, "details": details, "asyncAllowed": asyncAllowed})
	logger.Info("start")
	defer logger.Info("end")

	// Use async to process blockchain provision
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.client.azureRESTClient.resourceConfig.ResourceGroupName = instanceID
	err := b.client.Create(instanceID)
	if err != nil {
		logger.Error("create-blockchain-service", err)
		return brokerapi.ProvisionedServiceSpec{}, err
	}
	lastOperation := instanceID
	if err != nil {
		logger.Error("get-status-url", err)
		return brokerapi.ProvisionedServiceSpec{}, err
	}

	return brokerapi.ProvisionedServiceSpec{IsAsync: true, OperationData: "provision:" + lastOperation}, nil
}

func (b *ServiceBroker) Bind(context context.Context, instanceID string, bindingID string, details brokerapi.BindDetails) (_ brokerapi.Binding, e error) {
	logger := b.logger.Session("bind", lager.Data{"instanceID": instanceID})
	logger.Info("start")
	defer logger.Info("end")

	return brokerapi.Binding{}, nil
}

func (b *ServiceBroker) Update(context context.Context, instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.UpdateServiceSpec, error) {
	panic("not implemented")
}

func (b *ServiceBroker) Unbind(context context.Context, instanceID string, bindingID string, details brokerapi.UnbindDetails) (e error) {
	logger := b.logger.Session("unbind")
	logger.Info("start")
	defer logger.Info("end")

	return nil
}

func (b *ServiceBroker) Deprovision(context context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (_ brokerapi.DeprovisionServiceSpec, err error) {
	logger := b.logger.Session("deprovision").WithData(lager.Data{
		"instanceID":   instanceID,
		"details":      details,
		"asyncAllowed": asyncAllowed,
	})
	logger.Info("start")
	defer logger.Info("end")

	b.mutex.Lock()
	defer b.mutex.Unlock()

	client := *(b.client.azureRESTClient)
	client.resourceConfig.ResourceGroupName = instanceID
	exist, err := client.GroupExist()
	if err != nil {
		return brokerapi.DeprovisionServiceSpec{}, err
	}
	if !exist {
		return brokerapi.DeprovisionServiceSpec{}, nil
	}

	_, err = client.DeleteGroup()
	if err != nil {
		return brokerapi.DeprovisionServiceSpec{}, err
	}
	return brokerapi.DeprovisionServiceSpec{IsAsync: false, OperationData: "deprovision:" + instanceID}, nil
}
