# Azure Blockchain Service Broker

This is Service Broker that helps developers deploy Azure Ethereum Template to cloud foundry.

# Build AzureBlockchainBroker

```bash
git clone https://github.com/zeqing-guo/AzureBlockchainBroker.git
cd AzureBlockchainBroker
go get github.com/tools/godep
godep get .
GOOS=linux GOARCH=amd64 go build -o bin/AzureBlockchainBroker
```

# Configuration of AzureBlockchainBroker

To start AzureBlockchainBroker, all configurations must start with `--`. Please reference [Procflie](./Procflie).

- Environment variables for Broker
  - USERNAME: [REQUIRED] - Username for your broker.
  - PASSWORD: [REQUIRED] - Password for your broker.
- Configurations for Broker
  - listenAddr: (optional) - `host:port` to serve service broker API. Default value is `0.0.0.0:9000`. You must use the environment valriable `$PORT` if you deploy broker as a Cloud Foundry application. Please reference [here](https://docs.run.pivotal.io/devguide/deploy-apps/environment-variable.html#PORT).
  - serviceName: (optional) - name of the service to register with cloud controller. Default value is `azureblockchain`
  - serviceID: (optional) - ID of the service to register with cloud controller. Default value is `abb90071-f3e2-4a31-99f0-fc5d552dbbba`
- Configurations for Azure
  - environment: [REQUIRED] - The environment for Azure Management Service. Allowed values: `AzureCloud`, `AzureChinaCloud`, `AzureUSGovernment` or `AzureGermanCloud`. Default value is `AzureCloud`.
  - tenantID: [REQUIRED] - The tenant id for your service principal.
  - clientID: [REQUIRED] - The client id for your service principal.
  - clientSecret: [REQUIRED] - The client secret for your service principal.
  - subscriptionID: [REQUIRED] - The Azure Subscription id to use for storage accounts.
  - resourceGroupName: [REQUIRED] - The resource group name to use for storage accounts.
  - location: [REQUIRED] - The location to use for creating storage accounts.

  **NOTE:**

  - Please see more details about how to create a service principal [here](https://github.com/cloudfoundry-incubator/bosh-azure-cpi-release/blob/master/docs/get-started/create-service-principal.md).
  - `PORT` in [Procfile](./Procfile) will be allocated dynamically by Cloud Foundry runtime.
- Configurations for Blockchain Template
  - namePrefix: [REQUIRED] - String used as a base for naming resources.
  - adminUsername: (optional) - Administrator username of easch deployed VM.
  - adminPassword: [REQUIRED] - Administrator password for each deployed VM.
  - ethereumAccountPsswd: [REQUIRED] - Password used to secure the default Ethereum account that will be generated.
  - ethereumAccountPassphrase: [REQUIRED] - Password used to generate the private key associated with the default Ethereum account that is generated.
  - ethereumNetworkID: (optional) - Private Ethereum network ID to which to connect. Default value is 553289.
  - numConsortiumMembers: (optional) - Number of members within the network. The default value is 2.
  - numMiningNodesPerMember: (optional) - Number of mining nodes to create for each consortium member. The default value is 1.
  - mnNodeVMSize: (optional) - Size of the virtual machine used for mining nodes.
  - numTXNodes: (optional) - Number of load balanced transaction nodes. The default value is 1.
  - txNodeVMSize: (optional) - Size of the virtual machine for transaction nodes.
