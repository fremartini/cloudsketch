package private_dns_zone

import (
	"cloudsketch/internal/list"
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/models"
	"cloudsketch/internal/providers/azure/types"
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) GetResource(ctx *azContext.Context) ([]*models.Resource, error) {
	clientFactory, err := armprivatedns.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	dnsZone, err := clientFactory.NewPrivateZonesClient().Get(context.Background(), ctx.ResourceGroupName, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	resource := &models.Resource{
		Id:        *dnsZone.ID,
		Name:      *dnsZone.Name,
		Type:      *dnsZone.Type,
		DependsOn: []string{},
	}

	resources := []*models.Resource{resource}

	records, err := getRecordSet(clientFactory, ctx, *dnsZone.ID)

	if err != nil {
		return nil, err
	}

	resources = append(resources, records...)

	vnetLinks, err := getVnetLinks(clientFactory, ctx, *dnsZone.Name)

	if err != nil {
		return nil, err
	}

	resource.DependsOn = append(resource.DependsOn, vnetLinks...)

	return resources, nil
}

func getVnetLinks(clientFactory *armprivatedns.ClientFactory, ctx *azContext.Context, dnsZoneName string) ([]string, error) {
	pager := clientFactory.NewVirtualNetworkLinksClient().NewListPager(ctx.ResourceGroupName, dnsZoneName, nil)

	var links []*armprivatedns.VirtualNetworkLink
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		if resp.VirtualNetworkLinkListResult.Value != nil {
			links = append(links, resp.VirtualNetworkLinkListResult.Value...)
		}
	}

	return list.Map(links, func(l *armprivatedns.VirtualNetworkLink) string {
		return *l.Properties.VirtualNetwork.ID
	}), nil
}

func getRecordSet(clientFactory *armprivatedns.ClientFactory, ctx *azContext.Context, dnsZoneId string) ([]*models.Resource, error) {
	client := clientFactory.NewRecordSetsClient()

	pager := client.NewListPager(ctx.ResourceGroupName, ctx.ResourceName, nil)

	var records []*armprivatedns.RecordSet
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		if resp.RecordSetListResult.Value != nil {
			records = append(records, resp.RecordSetListResult.Value...)
		}
	}

	/*
		Microsoft.Network/privateDnsZones/A
		Microsoft.Network/privateDnsZones/SOA
		Microsoft.Network/privateDnsZones/CNAME
	*/

	// only A record contains IP addresses
	blacklist := []string{"Microsoft.Network/privateDnsZones/SOA", "Microsoft.Network/privateDnsZones/CNAME"}

	records = list.Filter(records, func(record *armprivatedns.RecordSet) bool {
		return !list.Contains(blacklist, func(blacklistItem string) bool {
			return *record.Type == blacklistItem
		})
	})

	resources := list.Map(records, func(record *armprivatedns.RecordSet) *models.Resource {
		return &models.Resource{
			Id:        *record.ID,
			Name:      *record.Name,
			Type:      types.DNS_RECORD,
			DependsOn: []string{dnsZoneId},
			Properties: map[string]string{
				"target": *record.Properties.ARecords[0].IPv4Address,
			},
		}
	})

	return resources, nil
}

func getRecordsInZone(dnsZoneId string, resources []*models.Resource) []*models.Resource {
	dnsRecords := list.Filter(resources, func(resource *models.Resource) bool {
		return resource.Type == types.DNS_RECORD
	})

	dnsRecords = list.Filter(dnsRecords, func(dnsRecord *models.Resource) bool {
		return list.Contains(dnsRecord.DependsOn, func(dependencyId string) bool {
			return dependencyId == dnsZoneId
		})
	})

	return dnsRecords
}

func (h *handler) PostProcess(resource *models.Resource, resources []*models.Resource) {
	recordsInDnsZone := getRecordsInZone(resource.Id, resources)

	for _, dnsRecord := range recordsInDnsZone {
		target, ok := dnsRecord.Properties["target"]

		if !ok {
			return
		}

		// attempt to find the resource with the target IP
		resourceWithIp := list.FirstOrDefault(resources, nil, func(nic *models.Resource) bool {
			ip, ok := nic.Properties["ip"]

			if !ok {
				return false
			}

			return ip == target
		})

		if resourceWithIp == nil {
			// unable to find matching IP
			return
		}

		// the resource that has the IP is likely a NIC. Search the attachedTo property
		attachedTo, ok := resourceWithIp.Properties["attachedTo"]

		if !ok {
			return
		}

		attachedToResource := list.FirstOrDefault(resources, nil, func(resource *models.Resource) bool {
			return attachedTo == strings.ToLower(resource.Id)
		})

		if attachedToResource == nil {
			return
		}

		dnsRecord.DependsOn = append(dnsRecord.DependsOn, attachedToResource.Id)
	}
}
