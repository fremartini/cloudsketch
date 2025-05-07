package providers

type Resource struct {
	Id, Type, Name string
	DependsOn      []string
	Properties     map[string][]string
}
