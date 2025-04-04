package models

type Resource struct {
	Id, Type, Name string
	DependsOn      []string
	Properties     map[string]any
}

func (r *Resource) GetLinkOrDefault() *string {
	link, ok := r.Properties["link"]
	if ok {
		var str = link.(string)
		return &str
	}

	return nil
}
