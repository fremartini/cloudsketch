package private_dns_zone

import (
	"cloudsketch/internal/az"
	"cloudsketch/internal/list"
	azContext "cloudsketch/internal/providers/azure/context"
	"cloudsketch/internal/providers/azure/types"
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
)

type handler struct{}

func New() *handler {
	return &handler{}
}

func (h *handler) Handle(ctx *azContext.Context) ([]*az.Resource, error) {
	clientFactory, err := armprivatedns.NewClientFactory(ctx.SubscriptionId, ctx.Credentials, nil)

	if err != nil {
		return nil, err
	}

	client := clientFactory.NewPrivateZonesClient()

	dns_zone, err := client.Get(context.Background(), ctx.ResourceGroup, ctx.ResourceName, nil)

	if err != nil {
		return nil, err
	}

	resource := &az.Resource{
		Id:            *dns_zone.ID,
		Name:          *dns_zone.Name,
		Type:          *dns_zone.Type,
		ResourceGroup: ctx.ResourceGroup,
		DependsOn:     []string{},
	}

	resources := []*az.Resource{resource}

	records, err := getRecordSet(clientFactory, ctx, dns_zone.ID)

	if err != nil {
		return nil, err
	}

	resources = append(resources, records...)

	return resources, nil
}

func getRecordSet(clientFactory *armprivatedns.ClientFactory, ctx *azContext.Context, dnsZoneId *string) ([]*az.Resource, error) {
	client := clientFactory.NewRecordSetsClient()

	pager := client.NewListPager(ctx.ResourceGroup, ctx.ResourceName, nil)

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

	resources := list.Map(records, func(record *armprivatedns.RecordSet) *az.Resource {
		return &az.Resource{
			Id:            *record.ID,
			Name:          *record.Name,
			Type:          types.DNS_RECORD,
			ResourceGroup: ctx.ResourceGroup,
			DependsOn:     []string{*dnsZoneId},
		}
	})

	// dont show @ records
	resources = list.Filter(resources, func(record *az.Resource) bool {
		return record.Name != "@"
	})

	return resources, nil
}
