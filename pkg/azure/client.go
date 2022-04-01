package azure

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/go-autorest/autorest"
)

func AcquireToken(ctx context.Context, tenantID string) (string, func(tenantID, resource string) (*autorest.BearerAuthorizer, error), error) {
	clientID := os.Getenv("AZURE_CLIENT_ID")
	tokenFilePath := os.Getenv("AZURE_FEDERATED_TOKEN_FILE")
	clientCert := os.Getenv("AZURE_CLIENT_CERTIFICATE")
	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")
	appIDURI := os.Getenv("AZURE_APP_ID_URI")

	if clientID == "" {
		return "unknown", nil, fmt.Errorf("missing client id")
	}

	if clientSecret != "" {
		return "token", AcquireTokenClientSecret(ctx, tenantID, appIDURI), nil
	} else if clientCert != "" {
		return "token", AcquireTokenClientCertificate(ctx, tenantID, appIDURI), nil
	} else if tokenFilePath != "" {
		return "callack", AcquireTokenFederatedToken(ctx, tenantID), nil
	} else {
		return "unknown", nil, fmt.Errorf("unable to determine method to get authentication")
	}
}
