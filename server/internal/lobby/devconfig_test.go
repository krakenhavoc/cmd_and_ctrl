package lobby

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/shoplinks"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/appenv"
)

func TestClientConfigProdReportsNoFeatures(t *testing.T) {
	c := Config{Env: appenv.EnvProd, Features: appenv.LoadFeatures(appenv.EnvProd)}
	rec := httptest.NewRecorder()
	if err := clientConfig(c, rec, httptest.NewRequest(http.MethodGet, "/config", nil)); err != nil {
		t.Fatalf("clientConfig: %v", err)
	}
	var body clientConfigResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Env != "prod" {
		t.Errorf("env = %q, want prod", body.Env)
	}
	if body.Features != (appenv.Features{}) {
		t.Errorf("features = %+v, want all false", body.Features)
	}
}

// ADR 0179: no shop configured serves an empty list, never null, and a
// configured one is served on any deployment, production included.
func TestClientConfigServesShops(t *testing.T) {
	c := Config{Env: appenv.EnvProd, Features: appenv.LoadFeatures(appenv.EnvProd)}
	rec := httptest.NewRecorder()
	if err := clientConfig(c, rec, httptest.NewRequest(http.MethodGet, "/config", nil)); err != nil {
		t.Fatalf("clientConfig: %v", err)
	}
	if !strings.Contains(rec.Body.String(), `"shops":[]`) {
		t.Errorf("body = %s, want \"shops\":[]", rec.Body.String())
	}

	c.Shops = []shoplinks.Shop{{ID: "tcgplayer", Name: "TCGplayer", Template: "{url}&partner=c2"}}
	rec = httptest.NewRecorder()
	if err := clientConfig(c, rec, httptest.NewRequest(http.MethodGet, "/config", nil)); err != nil {
		t.Fatalf("clientConfig: %v", err)
	}
	var body clientConfigResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Shops) != 1 || body.Shops[0] != c.Shops[0] {
		t.Errorf("shops = %+v, want %+v", body.Shops, c.Shops)
	}
}

// #3558: supporter links ride on GET /config as "support": {} when none
// is set, and only the configured ones when some are.
func TestClientConfigServesSupportLinks(t *testing.T) {
	c := Config{Env: appenv.EnvProd, Features: appenv.LoadFeatures(appenv.EnvProd)}
	rec := httptest.NewRecorder()
	if err := clientConfig(c, rec, httptest.NewRequest(http.MethodGet, "/config", nil)); err != nil {
		t.Fatalf("clientConfig: %v", err)
	}
	if !strings.Contains(rec.Body.String(), `"support":{}`) {
		t.Errorf("body = %s, want \"support\":{}", rec.Body.String())
	}

	c.Support.Kofi = "https://ko-fi.com/cmdctrl"
	rec = httptest.NewRecorder()
	if err := clientConfig(c, rec, httptest.NewRequest(http.MethodGet, "/config", nil)); err != nil {
		t.Fatalf("clientConfig: %v", err)
	}
	if !strings.Contains(rec.Body.String(), `"support":{"kofi":"https://ko-fi.com/cmdctrl"}`) {
		t.Errorf("body = %s, want the ko-fi link and no patreon key", rec.Body.String())
	}
}

func TestClientConfigDevReportsFeatures(t *testing.T) {
	c := Config{Env: appenv.EnvDev, Features: appenv.LoadFeatures(appenv.EnvDev)}
	rec := httptest.NewRecorder()
	if err := clientConfig(c, rec, httptest.NewRequest(http.MethodGet, "/config", nil)); err != nil {
		t.Fatalf("clientConfig: %v", err)
	}
	var body clientConfigResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Env != "dev" {
		t.Errorf("env = %q, want dev", body.Env)
	}
	if !body.Features.CardSpawn || !body.Features.SeatSwap {
		t.Errorf("features = %+v, want dev defaults on", body.Features)
	}
}

// The security-relevant test: a dev route is a 404 in prod even
// though the handler underneath would happily serve it.
func TestRequireDev(t *testing.T) {
	reached := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	requireDev(Config{Env: appenv.EnvProd}, inner).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dev/x", nil))
	if reached {
		t.Fatal("prod reached the dev handler")
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("prod status = %d, want 404 (403 would confirm the route exists)", rec.Code)
	}

	reached = false
	rec = httptest.NewRecorder()
	requireDev(Config{Env: appenv.EnvDev}, inner).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dev/x", nil))
	if !reached || rec.Code != http.StatusOK {
		t.Fatalf("dev did not reach handler: reached=%v status=%d", reached, rec.Code)
	}
}

func TestRequireDevFeature(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	pick := func(f appenv.Features) bool { return f.CardSpawn }

	// Dev deployment with the individual feature switched off.
	rec := httptest.NewRecorder()
	c := Config{Env: appenv.EnvDev, Features: appenv.Features{SeatSwap: true}}
	requireDevFeature(c, pick, inner).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dev/x", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("disabled feature status = %d, want 404", rec.Code)
	}

	rec = httptest.NewRecorder()
	c = Config{Env: appenv.EnvDev, Features: appenv.Features{CardSpawn: true}}
	requireDevFeature(c, pick, inner).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dev/x", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("enabled feature status = %d, want 200", rec.Code)
	}
}
