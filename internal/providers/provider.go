package providers

type Provider interface {
	FetchResources(subscriptionId string) ([]*Resource, string, error)
}
