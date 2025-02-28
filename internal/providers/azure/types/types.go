package types

const (
	APP_SERVICE_PLAN                      = "Microsoft.Web/serverFarms"
	APPLICATION_GATEWAY                   = "Microsoft.Network/applicationGateways"
	APPLICATION_INSIGHTS                  = "Microsoft.Insights/components"
	APPLICATION_SECURITY_GROUP            = "Microsoft.Network/applicationSecurityGroups"
	CONTAINER_REGISTRY                    = "Microsoft.ContainerRegistry/registries"
	DATA_FACTORY                          = "Microsoft.DataFactory/factories"
	DATA_FACTORY_INTEGRATION_RUNTIME      = "Microsoft.DataFactory/factories/integrationruntimes"
	DATA_FACTORY_MANAGED_PRIVATE_ENDPOINT = "Microsoft.DataFactory/factories/managedvirtualnetworks/managedprivateendpoints"
	DATABRICKS_WORKSPACE                  = "Microsoft.Databricks/workspaces"
	DNS_RECORD                            = "custom/privateDnsZones/record"
	KEY_VAULT                             = "Microsoft.KeyVault/vaults"
	LOAD_BALANCER                         = "Microsoft.Network/loadBalancers"
	LOAD_BALANCER_FRONTEND                = "Microsoft.Network/loadBalancers/frontendIPConfigurations"
	LOG_ANALYTICS                         = "Microsoft.OperationalInsights/workspaces"
	NAT_GATEWAY                           = "Microsoft.Network/natGateways"
	NETWORK_INTERFACE                     = "Microsoft.Network/networkInterfaces"
	NETWORK_SECURITY_GROUP                = "Microsoft.Network/networkSecurityGroups"
	POSTGRES_SQL_SERVER                   = "Microsoft.DBforPostgreSQL/servers"
	PRIVATE_DNS_ZONE                      = "Microsoft.Network/privateDnsZones"
	PRIVATE_ENDPOINT                      = "Microsoft.Network/privateEndpoints"
	PRIVATE_LINK_SERVICE                  = "Microsoft.Network/privateLinkServices"
	PUBLIC_IP_ADDRESS                     = "Microsoft.Network/publicIPAddresses"
	ROUTE_TABLE                           = "Microsoft.Network/routeTables"
	SQL_DATABASE                          = "Microsoft.Sql/servers/databases"
	SQL_SERVER                            = "Microsoft.Sql/servers"
	STORAGE_ACCOUNT                       = "Microsoft.Storage/storageAccounts"
	SUBNET                                = "Microsoft.Network/virtualNetworks/subnets"
	SUBSCRIPTION                          = "Microsoft.Subscription"
	VIRTUAL_MACHINE                       = "Microsoft.Compute/virtualMachines"
	VIRTUAL_MACHINE_SCALE_SET             = "Microsoft.Compute/virtualMachineScaleSets"
	VIRTUAL_NETWORK                       = "Microsoft.Network/virtualNetworks"
	WEB_SITES                             = "Microsoft.Web/sites"

	APP_SERVICE  = "appservice"
	FUNCTION_APP = "functionapp"
	LOGIC_APP    = "logicapp"
)
