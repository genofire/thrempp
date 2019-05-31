package threema

import "dev.sum7.eu/genofire/thrempp/component"

type Threema struct {
	component.Component
}

func NewThreema(config map[string]interface{}) (component.Component, error) {
	return &Threema{}, nil
}

func init() {
	component.AddComponent("threema", NewThreema)
}
