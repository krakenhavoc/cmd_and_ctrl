package legal_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// paid_for_targets_test.go -- #3546, ADR 0186 slice 4. "Target creature
// that saddled it this turn" (Giant Beaver) is one set everywhere: the
// engine's prompt, the enumerator the bot picks from, and the
// pick_target projection the client's picker highlights all walk the
// same clause, narrowed by TargetSpec.PaidForSource.

const oracleGiantBeaver = "2a4b42b3-6a95-41e1-9741-b35723d8da73"

func TestGiantBeaverEnumeratorAndViewOfferOnlyItsSaddlers(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	beaver := battlefieldCard(g, active, game.Card{
		Name: "Giant Beaver", TypeLine: "Creature — Beaver Mount", ManaCost: "{3}{G}",
		Power: 4, Toughness: 4, OracleID: oracleGiantBeaver,
	})
	other := battlefieldCard(g, active, game.Card{
		Name: "Gilded Ghoda", TypeLine: "Creature — Horse Mount", ManaCost: "{1}{R}",
		Power: 2, Toughness: 2, OracleID: oracleGildedGhoda,
	})
	a := battlefieldCard(g, active, creature("Saddler A", "{1}{G}", 2, 2))
	b := battlefieldCard(g, active, creature("Saddler B", "{1}{G}", 2, 2))
	c := battlefieldCard(g, active, creature("Saddler C", "{1}{G}", 2, 2))
	idle := battlefieldCard(g, active, creature("Idle", "{1}{G}", 2, 2))
	advanceTo(t, g, game.StepPrecombatMain)

	saddle := func(mount uuid.UUID, index int, saddlers ...uuid.UUID) {
		t.Helper()
		if err := g.ActivateCatalogAbility(active.ID, mount, index, game.ActivateAbilityParams{CrewIDs: saddlers}); err != nil {
			t.Fatalf("saddle: %v", err)
		}
		for i := 0; i < 8 && len(g.StackMeta) > 0; i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatal(err)
			}
		}
	}
	saddle(other, 0, c)
	saddle(beaver, 0, a, b)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(beaver, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	for i := 0; i < 8 && !hasPickTarget(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if !hasPickTarget(g) {
		t.Fatal("the Beaver's attack trigger never asked for a target")
	}

	want := []string{a.String(), b.String()}
	slices.Sort(want)

	moves := legal.EnumerateFor(g, active.ID)
	if len(moves) == 0 {
		t.Fatal("no answers to the Beaver's prompt -- the seat is stuck")
	}
	var offered []string
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("offered %q while the target prompt is open", m.Label)
			continue
		}
		var p struct {
			Target struct {
				ID string `json:"id"`
			} `json:"target"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("params %s: %v", string(m.Params), err)
		}
		offered = append(offered, p.Target.ID)
	}
	slices.Sort(offered)
	if !slices.Equal(offered, want) {
		t.Errorf("enumerator offers %v, want %v (not %s, %s or %s)", offered, want, c, idle, other)
	}
	dispatchAll(t, g, active.ID, moves)

	v := protocol.ViewOfGameFor(g, active.ID.String())
	var wire []string
	for _, pc := range v.PendingChoices {
		if pc.Kind == string(game.PendingChoicePickTarget) && pc.PickTarget != nil {
			wire = append(wire, pc.PickTarget.Cards...)
		}
	}
	slices.Sort(wire)
	if !slices.Equal(wire, want) {
		t.Errorf("view pick_target.cards = %v, want %v", wire, want)
	}
}

func hasPickTarget(g *game.Game) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			return true
		}
	}
	return false
}
