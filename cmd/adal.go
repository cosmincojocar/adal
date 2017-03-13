package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/cosmincojocar/adal"
	"os/user"
)

const (
	deviceMode       = "device"
	clientSecretMode = "secret"
	clientCertMode   = "cert"

	activeDirectoryEndpoint = "https://login.microsoftonline.com/"
)

type option struct {
	name  string
	value string
}

var (
	mode     string
	resource string

	tenantID      string
	applicationID string

	applicationSecret string
	certificatePath   string

	tokenCachePath string
)

func checkMondatoryOptions(mode string, options ...option) {
	for _, option := range options {
		if strings.TrimSpace(option.value) == "" {
			log.Fatalf("Authentication mode '%s' requires mandatory option '%s'.", mode, option.name)
		}
	}
}

func defaultTokenCachePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	defaultTokenPath := usr.HomeDir + "/.adal/accessToken.json"
	return defaultTokenPath
}

func init() {
	flag.StringVar(&mode, "mode", "device", "authentication mode (device, secret, cert)")
	flag.StringVar(&resource, "resource", "", "resource for which the token is requested")
	flag.StringVar(&tenantID, "tenantId", "", "tenant id")
	flag.StringVar(&applicationID, "applicationId", "", "application id")
	flag.StringVar(&applicationSecret, "secret", "", "application secret")
	flag.StringVar(&certificatePath, "certificatePath", "", "path to pk12/PFC application certificate")
	flag.StringVar(&tokenCachePath, "tokenCachePath", defaultTokenCachePath(), "location of oath token cache")

	flag.Parse()

	switch mode = strings.TrimSpace(mode); mode {
	case clientSecretMode:
		checkMondatoryOptions(clientSecretMode,
			option{name: "resource", value: resource},
			option{name: "tenantId", value: tenantID},
			option{name: "applicationId", value: applicationID},
			option{name: "secret", value: applicationSecret},
		)
	case clientCertMode:
		checkMondatoryOptions(clientCertMode,
			option{name: "resource", value: resource},
			option{name: "tenantId", value: tenantID},
			option{name: "applicationId", value: applicationID},
			option{name: "certificatePath", value: certificatePath},
		)
	case deviceMode:
		checkMondatoryOptions(deviceMode,
			option{name: "resource", value: resource},
			option{name: "tenantId", value: tenantID},
			option{name: "applicationId", value: applicationID},
		)
	default:
		log.Fatalln("Authentication modes 'secret, 'cert' or 'device' are supported.")
	}
}

func acquireTokenClientSecretFlow(oauthConfig adal.OAuthConfig,
	appliationID string,
	applicationSecret string,
	resource string,
	callbakcs ...adal.TokenRefreshCallback) (*adal.ServicePrincipalToken, error) {

	spt, err := adal.NewServicePrincipalToken(
		oauthConfig,
		appliationID,
		applicationSecret,
		resource,
		callbakcs...)
	if err != nil {
		return nil, err
	}

	return spt, spt.Refresh()
}

func saveToken(spt adal.Token) error {
	if tokenCachePath != "" {
		err := adal.SaveToken(tokenCachePath, 0600, spt)
		if err != nil {
			return err
		}
		log.Printf("Acquired token was saved in '%s' file\n", tokenCachePath)
		return nil

	}
	return fmt.Errorf("empty path for token cache")
}

func main() {
	oauthConfig, err := adal.NewOAuthConfig(activeDirectoryEndpoint, tenantID)
	if err != nil {
		panic(err)
	}

	callback := func(token adal.Token) error {
		return saveToken(token)
	}

	log.Printf("Authenticating with mode '%s'\n", mode)
	switch mode {
	case clientSecretMode:
		_, err = acquireTokenClientSecretFlow(
			*oauthConfig,
			applicationID,
			applicationSecret,
			resource,
			callback)
	}

	if err != nil {
		log.Fatalf("Failed to acquire a token for resource %s. Error: %v", resource, err)
	}
}
