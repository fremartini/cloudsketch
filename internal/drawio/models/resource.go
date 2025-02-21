package models

type Resource struct {
	Id, Type, Name string
	DependsOn      []string
	Properties     map[string]string
}

func (r *Resource) GetLinkOrDefault() *string {
	link, ok := r.Properties["link"]
	if ok {
		return &link
	}

	return nil
}
