package protocol

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// legal_actions_idle_test.go — #1918. A card's digest entry carries
// cast_idle_hint only when EVERY cast move for it is idle: one cast that
// would do something (Counterflux's targeted half, with a spell to
// point at) and the card highlights as an ordinary castable card.

func idleCastMove(t *testing.T, src uuid.UUID, alt, hint string) legal.Move {
	t.Helper()
	p, err := json.Marshal(map[string]any{"instance_id": src.String(), "from_zone": "hand", "alternative_cost": alt})
	if err != nil {
		t.Fatal(err)
	}
	return legal.Move{Type: legal.TypeCastSpell, Kind: legal.KindCast, Source: src, Params: p, IdleHint: hint}
}

func TestDigestCastIdleHintOnlyWhenEveryCastIsIdle(t *testing.T) {
	const hint = "Overloaded, this does nothing right now: there's no spell you don't control."
	onlyIdle := uuid.New()
	mixed := uuid.New()
	plain := uuid.New()
	moves := []legal.Move{
		idleCastMove(t, onlyIdle, "overload", hint),
		idleCastMove(t, mixed, "", ""),
		idleCastMove(t, mixed, "overload", hint),
		idleCastMove(t, plain, "", ""),
	}
	d := digestLegalMoves(moves)
	if d == nil {
		t.Fatal("no digest")
	}
	if got := d.Sources[onlyIdle.String()].CastIdleHint; got != hint {
		t.Errorf("every cast idle: cast_idle_hint %q, want %q", got, hint)
	}
	if got := d.Sources[mixed.String()].CastIdleHint; got != "" {
		t.Errorf("one cast does something: cast_idle_hint %q, want none", got)
	}
	if got := d.Sources[plain.String()].CastIdleHint; got != "" {
		t.Errorf("ordinary cast: cast_idle_hint %q, want none", got)
	}

	// Order must not matter: the idle cast first, then a live one.
	d = digestLegalMoves([]legal.Move{
		idleCastMove(t, mixed, "overload", hint),
		idleCastMove(t, mixed, "", ""),
	})
	if got := d.Sources[mixed.String()].CastIdleHint; got != "" {
		t.Errorf("idle first, live second: cast_idle_hint %q, want none", got)
	}

	raw, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var back struct {
		Sources map[string]map[string]json.RawMessage `json:"sources"`
	}
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	if _, ok := back.Sources[mixed.String()]["cast_idle_hint"]; ok {
		t.Errorf("cast_idle_hint must be omitted when empty: %s", raw)
	}
}
