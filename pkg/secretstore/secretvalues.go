package secretstore

import "encoding/json"

type SecretValue struct {
	Value          string
	PropertyValues map[string]string
	Labels         map[string]string
}

func (sv *SecretValue) ToString() string {
	if sv.Value != "" {
		return sv.Value
	}
	j, err := json.Marshal(sv)
	if err != nil {
		return "{}"
	}
	return string(j)
}
