package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// stack_ability_test.go — the enumerator half of #1865 (ADR 0106 §1
// amendment 2026-10-16): an activated ability of a SPELL that functions
// while it is on the stack. Lightning Storm's "Discard a land card: …"
// is any player's to activate, so it is a move for whoever holds
// priority and has a land card in hand, and the dispatcher accepts every
// such move (#544).

const oracleLightningStorm = "2c0fceb6-d92d-4f4a-85a2-0af5be5bba81"

// stormOnTheStack casts Lightning Storm from the active seat at `at`
// and returns its id. The caster keeps priority (CR 117.3c).
func stormOnTheStack(t *testing.T, g *game.Game, at *game.Player) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	storm := handCard(active, game.Card{Name: "Lightning Storm", TypeLine: "Instant", ManaCost: "{1}{R}{R}", OracleID: oracleLightningStorm})
	for i := 0; i < 3; i++ {
		battlefieldCard(g, active, basic("Mountain", "Mountain"))
	}
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(active.ID, storm, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: at.ID}}, Strict: true, AutoTap: true,
	}); err != nil {
		t.Fatalf("cast Lightning Storm: %v", err)
	}
	return storm
}

func TestStackAbilityIsOfferedToItsController(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	land := handCard(active, basic("Mountain", "Mountain"))
	handCard(active, creature("Not a land", "{1}", 1, 1))
	storm := stormOnTheStack(t, g, opp)

	moves := legal.EnumerateFor(g, active.ID)
	var acts []legal.Move
	for _, m := range moves {
		if m.Kind == legal.KindActivate && m.Source == storm {
			acts = append(acts, m)
		}
	}
	if len(acts) != 1 {
		t.Fatalf("the caster with one land card: want one Lightning Storm activation, got %v", labels(moves))
	}
	if !strings.Contains(string(acts[0].Params), land.String()) {
		t.Errorf("the activation should discard the land card %s: %s", land, acts[0].Params)
	}
	if !strings.HasPrefix(acts[0].Label, "Lightning Storm: ") {
		t.Errorf("label %q: the caster's own spell names no controller", acts[0].Label)
	}
	dispatchAll(t, g, active.ID, acts)
}

func TestStackAbilityIsOfferedToAnyPlayerWithALand(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	clearHand(opp)
	storm := stormOnTheStack(t, g, opp)
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	if !holdsPriority(g, opp.ID) {
		t.Fatal("the next seat should hold priority")
	}

	// No land card in hand: nothing to pay with, so no move.
	handCard(opp, creature("Not a land", "{1}", 1, 1))
	for _, m := range legal.EnumerateFor(g, opp.ID) {
		if m.Kind == legal.KindActivate && m.Source == storm {
			t.Fatalf("offered with no land card to discard: %q", m.Label)
		}
	}

	handCard(opp, basic("Forest", "Forest"))
	moves := legal.EnumerateFor(g, opp.ID)
	var acts []legal.Move
	for _, m := range moves {
		if m.Kind == legal.KindActivate && m.Source == storm {
			acts = append(acts, m)
		}
	}
	if len(acts) != 1 {
		t.Fatalf("a non-controller with a land card: want one activation, got %v", labels(moves))
	}
	if want := "Lightning Storm (controlled by " + active.Name + "): "; !strings.HasPrefix(acts[0].Label, want) {
		t.Errorf("label %q, want it to name the spell's controller (%q…)", acts[0].Label, want)
	}
	dispatchAll(t, g, opp.ID, acts)
}

// An ordinary instant has no ability that functions on the stack, so the
// walk offers nothing for it (the fast negative).
func TestOrdinarySpellOffersNoStackAbility(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	bolt := handCard(active, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt})
	handCard(active, basic("Mountain", "Mountain"))
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(active.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true, AutoTap: true}); err != nil {
		t.Fatal(err)
	}
	for _, m := range legal.EnumerateFor(g, active.ID) {
		if m.Kind == legal.KindActivate && m.Source == bolt {
			t.Errorf("a Lightning Bolt on the stack offered an activation: %q", m.Label)
		}
	}
}
