package lobby

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/budget"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
)

// botmodel_test.go: #2980's HTTP half. The kill switch is admin-only
// and takes effect at once; GET /bot/options carries the note while
// the model tiers are off.

func putJSON(t *testing.T, srv *httptest.Server, path, token, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, srv.URL+path, bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func newBotModelServer(t *testing.T, b BotModelBudget) (*httptest.Server, string, string) {
	t.Helper()
	l := newTestLobby(t)
	host := newFakeBotStatsHost()
	l.SetBotHost(host)
	a := newTestAuth()
	meta, _ := l.Create("FNM")
	srv := httptest.NewServer(Handler(Config{Lobby: l, Auth: a, Bots: host, BotBudget: b, Log: discardLogger()}))
	t.Cleanup(srv.Close)
	return srv, adminSession(t, a), playerSession(t, a, meta.ID)
}

func TestBotModelSwitchIsAdminOnly(t *testing.T) {
	b := budget.New(budget.DefaultLimits(), nil)
	srv, admin, player := newBotModelServer(t, b)

	for _, tc := range []struct {
		name, token string
		want        int
	}{{"anonymous", "", http.StatusUnauthorized}, {"seated player", player, http.StatusForbidden}} {
		resp := putJSON(t, srv, "/admin/bot/model", tc.token, `{"enabled":false}`)
		resp.Body.Close()
		if resp.StatusCode != tc.want {
			t.Errorf("%s PUT: %d, want %d", tc.name, resp.StatusCode, tc.want)
		}
		resp = doGet(t, srv, "/admin/bot/model", tc.token)
		resp.Body.Close()
		if resp.StatusCode != tc.want {
			t.Errorf("%s GET: %d, want %d", tc.name, resp.StatusCode, tc.want)
		}
	}
	if !b.Status().Enabled {
		t.Fatal("a refused request flipped the switch")
	}

	resp := putJSON(t, srv, "/admin/bot/model", admin, `{"enabled":false}`)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("admin PUT: %d %s", resp.StatusCode, body)
	}
	var st budget.Status
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		t.Fatal(err)
	}
	if st.Enabled || st.State != budget.StateDisabled {
		t.Fatalf("response = %+v, want the switch off", st)
	}
	if b.Status().Enabled {
		t.Fatal("the budget still says enabled")
	}
}

func TestBotModelSwitchValidatesItsBody(t *testing.T) {
	srv, admin, _ := newBotModelServer(t, budget.New(budget.DefaultLimits(), nil))
	for _, body := range []string{`{}`, `{"enabled":"no"}`, `{"enabled":false,"extra":1}`, ``} {
		resp := putJSON(t, srv, "/admin/bot/model", admin, body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("body %q: %d, want 400 (a missing field must never read as off)", body, resp.StatusCode)
		}
	}
}

func TestBotModelStatusReports503WithNoGuard(t *testing.T) {
	srv, admin, _ := newBotModelServer(t, nil)
	resp := doGet(t, srv, "/admin/bot/model", admin)
	resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", resp.StatusCode)
	}
}

func TestBotModelStatusReadsTheCounters(t *testing.T) {
	b := budget.New(budget.Limits{DailyTokens: 1000, MonthlyTokens: 5000}, nil, budget.WithWeights(budget.FlatWeights()))
	b.Record(model.Usage{InputTokens: 300, OutputTokens: 100})
	srv, admin, _ := newBotModelServer(t, b)
	resp := doGet(t, srv, "/admin/bot/model", admin)
	defer resp.Body.Close()
	var st budget.Status
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		t.Fatal(err)
	}
	if st.DailyTokens != 400 || st.DailyLimit != 1000 || st.MonthlyLimit != 5000 || st.State != budget.StateOK {
		t.Fatalf("status = %+v", st)
	}
}

func botOptionsBody(t *testing.T, srv *httptest.Server, token string) botOptionsResponse {
	t.Helper()
	resp := doGet(t, srv, "/bot/options", token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /bot/options: %d", resp.StatusCode)
	}
	var out botOptionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestBotOptionsCarriesTheNoteWhileTheModelTiersAreOff(t *testing.T) {
	b := budget.New(budget.Limits{DailyTokens: 100}, nil, budget.WithWeights(budget.FlatWeights()))
	srv, _, player := newBotModelServer(t, b)

	if got := botOptionsBody(t, srv, player); got.ModelNote != "" {
		t.Fatalf("under the cap: model_note = %q, want none", got.ModelNote)
	}

	b.Record(model.Usage{InputTokens: 100})
	got := botOptionsBody(t, srv, player)
	if !strings.Contains(got.ModelNote, "daily model budget is used up") || !strings.Contains(got.ModelNote, "UTC") {
		t.Fatalf("over the cap: model_note = %q, want the daily-cap note with its reset time", got.ModelNote)
	}

	_ = b.SetEnabled(false)
	if got := botOptionsBody(t, srv, player); !strings.Contains(got.ModelNote, "switched model calls off") {
		t.Fatalf("switched off: model_note = %q", got.ModelNote)
	}
}
