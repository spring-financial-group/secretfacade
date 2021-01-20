package secretstore

import (
	"encoding/json"

	"github.com/imdario/mergo"
)

type SecretValue struct {
	Value          string
	PropertyValues map[string]string
	Labels         map[string]string
	Overwrite      bool
}

func (sv *SecretValue) ToString() string {
	if sv.Value != "" {
		return sv.Value
	}
	j, err := json.Marshal(sv.PropertyValues)
	if err != nil {
		return "{}"
	}
	return string(j)
}

func (sv *SecretValue) MergeExistingSecret(existing map[string]string) string {
	if existing == nil {
		return sv.ToString()
	}
	if sv.Value != "" {
		return sv.Value
	}
	err := mergo.Merge(&existing, sv.PropertyValues, mergo.WithOverride)

	j, err := json.Marshal(existing)
	if err != nil {
		return "{}"
	}
	return string(j)

}
