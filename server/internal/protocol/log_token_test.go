package protocol

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_token_test.go — #4441, ADR 0209 §3: the client's token_create cue
// reads what the public log already says about a created token. Nothing
// on the wire is new; this pins that every viewer gets one `token` entry
// per token, naming the token, and that tokens one effect creates share a
// batch, which is how the cue plays one sound for them.
func TestCreatedTokensAreLoggedOnePerTokenInOneBatch(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID
	var made []uuid.UUID
	g.WithWriteLock(func() {
		ids, err := g.CreateTokensForEffect(me, game.Card{
			Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1,
		}, 3, game.TokenEntryOptions{})
		if err != nil || len(ids) != 3 {
			t.Fatalf("create three tokens: %v %v", ids, err)
		}
		made = ids
	})

	// An opponent sees them as well as their controller does.
	log := ViewOfGameFor(g, g.Seats[1].ID.String()).Log
	want := map[string]bool{}
	for _, id := range made {
		want[id.String()] = true
	}
	var tokens []LogEvent
	for _, e := range log {
		if !want[e.CardID] {
			continue
		}
		if e.Kind == LogZone {
			// A token comes from no zone: its creation is the
			// announcement, never a zone move as well (#762).
			t.Errorf("a created token also has a zone entry: %+v", e)
		}
		if e.Kind == LogToken {
			tokens = append(tokens, e)
		}
	}
	if len(tokens) != 3 {
		t.Fatalf("got %d token entries, want 3: %v", len(tokens), logKinds(log))
	}
	for i, e := range tokens {
		if e.CardID != made[i].String() {
			t.Errorf("token entry %d names %q, want %s (log order is creation order)", i, e.CardID, made[i])
		}
		if e.Batch == 0 || e.Batch != tokens[0].Batch {
			t.Errorf("batch %d, want one non-zero batch for the whole creation (%d)", e.Batch, tokens[0].Batch)
		}
	}

	// The names on the wire.
	buf, err := json.Marshal(tokens[0])
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(buf, &raw); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]any{"kind": "token", "card_id": made[0].String(), "batch": float64(tokens[0].Batch)} {
		if raw[key] != want {
			t.Errorf("wire field %q = %v, want %v (entry %s)", key, raw[key], want, buf)
		}
	}
}
