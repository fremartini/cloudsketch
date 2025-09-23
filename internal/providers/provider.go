package providers

type Provider interface {
	FetchResources(input string) ([]*Resource, string, error)
}
