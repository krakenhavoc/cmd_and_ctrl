package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// divide_shield_test.go — ADR 0108 §7 decision 6 (#1904): a bot owing a
// divide_shield prompt is offered answers the engine accepts (#544), the
// protective division first ("the heuristic shields the player first"),
// and the engine's order always.
func TestDivideShieldOffersTheProtectiveDivisionFirst(t *testing.T) {
	g := newTable(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := game.Card{InstanceID: uuid.New(), Name: "Pyromancer", TypeLine: "Creature — Test", Power: 2, Toughness: 2,
		Owner: opp.ID, Controller: opp.ID}
	bear := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID}
	g.Battlefield.PushTop(src)
	g.Battlefield.PushTop(bear)
	g.WithWriteLock(func() {
		ref, zone, ok := g.DamageSourceRefLocked(src.InstanceID)
		if !ok {
			t.Fatal("setup: no source")
		}
		g.PreventDamageFromSourceThisTurnForEffect(game.DamageShield{
			Controller: me.ID, Source: ref, SourceZone: zone, ProtectPlayer: me.ID,
			ProtectTypes: []string{"creature"}, Amount: 2, Label: "Refraction Trap",
		})
		_ = g.DamageInstanceForEffect(func() error {
			if err := g.DealDamageToCreatureForEffect(src.InstanceID, bear.InstanceID, 2); err != nil {
				return err
			}
			return g.DealDamageToPlayerForEffect(src.InstanceID, me.ID, 3)
		})
	})
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceDivideShield {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatal("setup: no divide_shield prompt")
	}

	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) != 2 {
		t.Fatalf("enumerated %d answers, want 2: %v", len(moves), labels(moves))
	}
	if !moves[1].AlwaysLegal {
		t.Error("the engine's order is not marked always-legal")
	}
	var first struct {
		Distribution map[string]int `json:"distribution"`
	}
	if err := json.Unmarshal(moves[0].Params, &first); err != nil {
		t.Fatal(err)
	}
	for _, en := range prompt.DivideShield.Entries {
		want := 0
		if en.Target == me.ID {
			want = 2
		}
		if got := first.Distribution[en.ID.String()]; got != want {
			t.Errorf("protective share for %s = %d, want %d (the player first)", en.TargetName, got, want)
		}
	}
	dispatchAll(t, g, me.ID, moves)
}
