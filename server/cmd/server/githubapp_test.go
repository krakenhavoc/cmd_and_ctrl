package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
	"testing"
)

func TestGitHubAppFromEnv(t *testing.T) {
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)})
	env := func(m map[string]string) func(string) string {
		return func(k string) string { return m[k] }
	}

	if src, err := githubAppFromEnv(env(nil)); src != nil || err != nil {
		t.Errorf("none set: %v %v, want no app and no error", src, err)
	}
	full := map[string]string{
		envGitHubAppID:             "7",
		envGitHubAppInstallationID: "42",
		envGitHubAppPrivateKey:     base64.StdEncoding.EncodeToString(keyPEM),
	}
	if src, err := githubAppFromEnv(env(full)); src == nil || err != nil {
		t.Errorf("all set: %v %v", src, err)
	}

	for name, tweak := range map[string]func(m map[string]string){
		"no installation": func(m map[string]string) { delete(m, envGitHubAppInstallationID) },
		"id not a number": func(m map[string]string) { m[envGitHubAppID] = "seven" },
		"zero install":    func(m map[string]string) { m[envGitHubAppInstallationID] = "0" },
		"bad key":         func(m map[string]string) { m[envGitHubAppPrivateKey] = "SECRETVALUE" },
	} {
		m := map[string]string{}
		for k, v := range full {
			m[k] = v
		}
		tweak(m)
		src, err := githubAppFromEnv(env(m))
		if src != nil || err == nil {
			t.Errorf("%s: %v %v, want an error", name, src, err)
			continue
		}
		if strings.Contains(err.Error(), "SECRETVALUE") {
			t.Errorf("%s: the error quotes the key: %v", name, err)
		}
	}
}
