package gcpiam

import (
	"context"

	"golang.org/x/oauth2/google"
)

func DefaultCredentials() (*google.Credentials, error) {
	adc, err := google.FindDefaultCredentials(context.TODO(), "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, err
	}
	return adc, nil
}
