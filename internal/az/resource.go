package az

// TODO: decouple this. ResourceGroup is not needed in the DrawIO domain
type Resource struct {
	Id, Type, Name, ResourceGroup string
	DependsOn                     []string
	Properties                    map[string]string
}
