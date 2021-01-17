package fake

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func (f FakeSecretStore) AssertHasValue(t *testing.T, location string, secretName string, secretKey string) {
	secret, err := f.GetSecret(location, secretName, secretKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, secret, "no value found for secret %s and property %s at location %s", secretName, secretKey, location)
}

func (f FakeSecretStore) AssertValueEquals(t *testing.T, location string, secretName string, secretKey string, expectedValue string) {
	secret, err := f.GetSecret(location, secretName, secretKey)
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, secret, "value does not match for secret %s and property %s at location %s, expected %s but got %s", secretName, secretKey, location, expectedValue, secret)
}
