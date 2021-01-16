package gcp

import (
	"context"

	"golang.org/x/oauth2/google"
)

func DefaultCredentials(ctx context.Context) (*google.Credentials, error) {
	adc, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, err
	}
	return adc, nil
}
