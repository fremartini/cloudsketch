package models

type Resource struct {
	Id, Type, Name, ResourceGroup string
	DependsOn                     []string
	Properties                    map[string]string
}
