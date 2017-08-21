export USERNAME=admin
export PASSWORD=admin
export PORT=9000
./AzureBlockchainBroker -listenAddr "127.0.0.1:9000" -tenantID "" -clientID "" -clientSecret "" -subscriptionID "" -location "southcentralus" -namePrefix "ethnet" -adminUsername "gethadmin" -adminPassword "" -ethereumAccountPsswd "" -ethereumAccountPassphrase "" -ethereumNetworkID 553289 -numConsortiumMembers 2 -numMiningNodesPerMember 1 -mnNodeVMSize "Standard_D1_v2" -numTXNodes 1 -txNodeVMSize "Standard_D1_v2"
