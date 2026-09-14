package lobby

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// bot_tiers_test.go covers what GET /bot/options says about the
// difficulty slider once a real policy factory is wired in, and what
// POST /games/{id}/seats/bot does with a tier this server cannot
// play.
//
// The two are one subject: the picker's availability flag and the
// endpoint's refusal have to agree, because the failure they exist to
// prevent is a seat labelled with a tier it is not playing.

// explainingBotHost is a host that can both offer tiers and say why
// it is withholding the others — *aiseat.Manager's shape, without the
// runners.
type explainingBotHost struct {
	*fakeBotHost
	reasons map[string]string
}

func (h *explainingBotHost) TierReason(tier string) string { return h.reasons[tier] }

func newExplainingHost(available []string, reasons map[string]string) *explainingBotHost {
	h := &explainingBotHost{fakeBotHost: newFakeBotHost(), reasons: reasons}
	h.tiers = available
	return h
}

func botOptionsStack(t *testing.T, host BotHost) (*httptest.Server, auth.Authenticator) {
	t.Helper()
	log := discardLogger()
	mgr := ws.NewRoomManager(log, "")
	l := NewLobby(mgr)
	l.SetBotHost(host)
	a := newTestAuth()
	srv := newTestServerWithConfig(t, Config{
		Lobby: l, Auth: a, AdminToken: "shared-admin-token",
		Cards: buildMinimalDeckIndex(t), Bots: host, BotDecks: fakeDeckSource{},
	})
	return srv, a
}

func readBotOptions(t *testing.T, srv *httptest.Server, a auth.Authenticator) botOptionsResponse {
	t.Helper()
	resp := doGet(t, srv, "/bot/options", adminSession(t, a))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("options: got %d", resp.StatusCode)
	}
	var out botOptionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}

// A server whose factory can build all four offers all four. This is
// the state a deployment with a model endpoint is in, and until S31's
// final wiring it was unreachable.
func TestBotOptionsOffersEveryTierAHostCanBuild(t *testing.T) {
	host := newExplainingHost([]string{"random", "heuristic", "assisted", "strong"}, nil)
	srv, a := botOptionsStack(t, host)

	out := readBotOptions(t, srv, a)
	if len(out.Tiers) != 4 {
		t.Fatalf("tiers: %+v", out.Tiers)
	}
	for _, ti := range out.Tiers {
		if !ti.Available {
			t.Errorf("tier %q unavailable although the host builds it", ti.Tier)
		}
		if ti.Reason != "" {
			t.Errorf("available tier %q carries a reason %q", ti.Tier, ti.Reason)
		}
		if strings.Contains(ti.Description, "sub-PR") {
			t.Errorf("tier %q still describes itself as unshipped: %q", ti.Tier, ti.Description)
		}
	}
}

// A server with no model endpoint offers the two free tiers, withholds
// the two that need a model, and says WHY — the reason is the thing
// that turns a greyed-out row into a configuration change.
func TestBotOptionsExplainsAWithheldModelTier(t *testing.T) {
	const why = "needs a model: set CMDCTRL_OPENAI_ENDPOINT or CMDCTRL_ANTHROPIC_API_KEY"
	host := newExplainingHost([]string{"random", "heuristic"}, map[string]string{
		string(aiseat.TierAssisted): why,
		string(aiseat.TierStrong):   why,
	})
	srv, a := botOptionsStack(t, host)

	for _, ti := range readBotOptions(t, srv, a).Tiers {
		switch ti.Tier {
		case aiseat.TierRandom, aiseat.TierHeuristic:
			if !ti.Available || ti.Reason != "" {
				t.Errorf("%q: available=%v reason=%q", ti.Tier, ti.Available, ti.Reason)
			}
		case aiseat.TierAssisted, aiseat.TierStrong:
			if ti.Available {
				t.Errorf("%q reported available on a server with no model endpoint", ti.Tier)
			}
			if ti.Reason != why {
				t.Errorf("%q reason = %q, want the host's explanation", ti.Tier, ti.Reason)
			}
		}
	}
}

// Asking for a tier this server cannot play is a 422 and NO SEAT. Not
// a 500, and — the point of the whole tier system — not a seat that
// quietly plays at a weaker tier under the name that was asked for.
func TestAddBotRefusesAnUnavailableTierWithoutDowngrading(t *testing.T) {
	host := newExplainingHost([]string{"random", "heuristic"}, map[string]string{
		string(aiseat.TierAssisted): "needs a model endpoint",
	})
	srv, a := botOptionsStack(t, host)
	token := adminSession(t, a)
	resp := postJSON(t, srv, "/games", token, createGameRequest{Name: "FNM"})
	defer resp.Body.Close()
	var meta GameMeta
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		t.Fatal(err)
	}

	bad := postJSON(t, srv, "/games/"+meta.ID.String()+"/seats/bot", token, addBotRequest{
		Tier: string(aiseat.TierAssisted), Deck: "test-mono-white",
	})
	defer bad.Body.Close()
	if bad.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("unavailable tier: got %d, want 422", bad.StatusCode)
	}
	if len(host.started) != 0 {
		t.Errorf("a refused tier started runners: %+v", host.started)
	}

	// And the seat was never created: a 422 that half-seated a bot
	// would be worse than either outcome.
	after := doGet(t, srv, "/games/"+meta.ID.String(), token)
	defer after.Body.Close()
	var got GameMeta
	if err := json.NewDecoder(after.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	for _, p := range got.Players {
		if p.IsBot {
			t.Errorf("a bot was seated despite the 422: %+v", p)
		}
	}

	// The tier the host DOES offer still works, so the refusal is
	// about availability and not about the endpoint being broken.
	ok := postJSON(t, srv, "/games/"+meta.ID.String()+"/seats/bot", token, addBotRequest{
		Tier: string(aiseat.TierHeuristic), Deck: "test-mono-white",
	})
	defer ok.Body.Close()
	if ok.StatusCode != http.StatusCreated {
		t.Fatalf("available tier: got %d, want 201", ok.StatusCode)
	}
	var seated addBotResponse
	if err := json.NewDecoder(ok.Body).Decode(&seated); err != nil {
		t.Fatal(err)
	}
	for _, p := range seated.Game.Players {
		if p.IsBot && p.BotTier != string(aiseat.TierHeuristic) {
			t.Errorf("seat records tier %q, want heuristic", p.BotTier)
		}
	}
}
