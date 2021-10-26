//go:build unit
// +build unit

package vaultiam_test

import (
	"os"
	"testing"

	"github.com/jenkins-x-plugins/secretfacade/pkg/iam/vaultiam"
	"github.com/stretchr/testify/assert"
)

const testToken = "token123"

var testcases = []struct {
	description string
	vaultToken  string
	expectError bool
}{
	{
		description: "Test case: Internal vault with no token set",
		vaultToken:  "",
		expectError: true,
	},
	{
		description: "Test case: Internal vault with token set",
		vaultToken:  testToken,
		expectError: false,
	},
}

func TestNewEnvironmentCreds(t *testing.T) {
	for _, tt := range testcases {
		t.Log(tt.description)
		os.Setenv("VAULT_TOKEN", tt.vaultToken)
		vaultcreds, err := vaultiam.NewEnvironmentCreds()

		if tt.expectError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		if tt.vaultToken != "" {
			assert.NotEmpty(t, vaultcreds.Token)
		} else {
			assert.Empty(t, vaultcreds.Token)
		}
	}
}
