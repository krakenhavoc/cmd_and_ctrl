package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/github"
)

// The GitHub App that files in-app reports as its bot (#3493). All
// three or none: the key is the app's private key, PEM or that PEM
// base64-encoded on one line (the env file holds single lines).
const (
	envGitHubAppID             = "CMDCTRL_GITHUB_APP_ID"
	envGitHubAppInstallationID = "CMDCTRL_GITHUB_APP_INSTALLATION_ID"
	envGitHubAppPrivateKey     = "CMDCTRL_GITHUB_APP_PRIVATE_KEY"
)

// githubAppFromEnv reads the app's settings. None set is (nil, nil): no
// app, and the PAT (if any) files instead. Some but not all, an id that
// is not a positive number, or a key that does not parse is an error, so
// a broken app config fails the boot rather than quietly filing as the
// PAT's owner. The error never quotes the key.
func githubAppFromEnv(getenv func(string) string) (*github.AppTokenSource, error) {
	vals := map[string]string{}
	var missing []string
	for _, k := range []string{envGitHubAppID, envGitHubAppInstallationID, envGitHubAppPrivateKey} {
		v := strings.TrimSpace(getenv(k))
		vals[k] = v
		if v == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) == 3 {
		return nil, nil
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("the GitHub App needs all three settings; missing %s", strings.Join(missing, ", "))
	}
	appID, err := strconv.ParseInt(vals[envGitHubAppID], 10, 64)
	if err != nil || appID <= 0 {
		return nil, fmt.Errorf("%s must be a positive number", envGitHubAppID)
	}
	installationID, err := strconv.ParseInt(vals[envGitHubAppInstallationID], 10, 64)
	if err != nil || installationID <= 0 {
		return nil, fmt.Errorf("%s must be a positive number", envGitHubAppInstallationID)
	}
	src, err := github.NewAppTokenSource(appID, installationID, []byte(vals[envGitHubAppPrivateKey]))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", envGitHubAppPrivateKey, err)
	}
	return src, nil
}
