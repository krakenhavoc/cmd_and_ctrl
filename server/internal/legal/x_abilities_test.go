package legal_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// x_abilities_test.go — the enumerator half of {X} on an activated
// ability. The invariant every case here is defending is #544's:
// the enumerator must offer exactly the moves the engine accepts,
// and the expansion budget must reach every arity rather than being
// spent on one.

const (
	oracleTreasureVault   = "3c43efd6-b1a8-452c-ae20-9a936c3340ab"
	oracleHelmOfObedience = "16cadebf-c484-41f8-9e38-5c2c528f5b54"
	oracleSoothsaying     = "517b702a-2c5c-40d7-825e-6c674019b298"
	oracleSmugglersCopter = "49136bdc-bc50-49a2-999a-1ef9c16ea130"
)

// xValueOf pulls x_value out of an activate_ability move's params.
func xValueOf(t *testing.T, m legal.Move) int {
	t.Helper()
	var p struct {
		XValue int `json:"x_value"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", string(m.Params), err)
	}
	return p.XValue
}

func activationsOf(moves []legal.Move, source uuid.UUID) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Kind == legal.KindActivate && m.Source == source {
			out = append(out, m)
		}
	}
	return out
}

// mana seeds n untapped colorless sources for a seat.
func mana(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n; i++ {
		battlefieldCard(g, p, game.Card{Name: fmt.Sprintf("Island %d", i), TypeLine: "Basic Land — Island"})
	}
}

func TestActivatedXOffersTheLargestAffordableValue(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	vault := battlefieldCard(g, active, game.Card{
		Name: "Treasure Vault", TypeLine: "Artifact Land", OracleID: oracleTreasureVault,
	})
	// Seven other sources. "{X}{X}" is two slots, so seven mana buys
	// X=3 and no more. The Vault itself is NOT a source here: its
	// cost includes {T}, so it cannot tap for mana and pay its own
	// tap cost — which is exactly what the engine refuses, and the
	// enumerator has to agree.
	mana(g, active, 7)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, vault)
	if len(acts) != 1 {
		t.Fatalf("want exactly one Treasure Vault activation (one X, not a range), got %d: %v", len(acts), labels(moves))
	}
	if got := xValueOf(t, acts[0]); got != 3 {
		t.Errorf("X = %d, want 3 (seven mana over two {X} slots)", got)
	}
	if !strings.Contains(acts[0].Label, "X=3") {
		t.Errorf("label should name the chosen X: %q", acts[0].Label)
	}
	// The soundness invariant: the engine accepts what we offered.
	dispatchAll(t, g, active.ID, moves)
}

func TestActivatedXCannotTapItsOwnSourceForMana(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	vault := battlefieldCard(g, active, game.Card{
		Name: "Treasure Vault", TypeLine: "Artifact Land", OracleID: oracleTreasureVault,
	})
	// Three other sources. "{X}{X}" is two slots, so three outside
	// mana buys X=1. If the Vault could tap for {C} as well the
	// enumerator would read four and offer X=2 — an activation the
	// engine refuses, because the same permanent cannot pay both the
	// {T} and part of the {X}{X}.
	//
	// Three rather than one since #810: X=0 makes no Treasures and
	// sacrifices the land for nothing, so it is not offered at all
	// and a one-source board proves nothing about the exclusion.
	mana(g, active, 3)
	advanceTo(t, g, game.StepPrecombatMain)

	acts := activationsOf(legal.EnumerateFor(g, active.ID), vault)
	if len(acts) != 1 {
		t.Fatalf("want one activation, got %d", len(acts))
	}
	if got := xValueOf(t, acts[0]); got != 1 {
		t.Errorf("X = %d, want 1 — three outside sources fund {X}{X} at X=1 and no further", got)
	}
	dispatchAll(t, g, active.ID, legal.EnumerateFor(g, active.ID))
}

// The other half of the same board: with one outside source the
// Vault's only legal announcement is X=0, which creates no Treasures
// and still sacrifices the land, so the ability is not a move at all
// (#810, CR 732.2a).
func TestActivatedXIsNotOfferedWhenOnlyXZeroIsPayable(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	vault := battlefieldCard(g, active, game.Card{
		Name: "Treasure Vault", TypeLine: "Artifact Land", OracleID: oracleTreasureVault,
	})
	mana(g, active, 1)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	if acts := activationsOf(moves, vault); len(acts) != 0 {
		t.Fatalf("one mana buys no Treasures, so the Vault has no activation worth offering, got %v", labels(acts))
	}
	// The Vault's own mana ability is untouched: a zero-effect {X}
	// activation is not a move, tapping for colourless still is.
	if !hasLabel(moves, "Treasure Vault: Add") {
		t.Errorf("the Vault's mana ability went missing with its {X} ability: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}

func TestHelmOfObedienceIsNotOfferedBelowItsPrintedFloor(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	helm := battlefieldCard(g, active, game.Card{
		Name: "Helm of Obedience", TypeLine: "Artifact", ManaCost: "{4}", OracleID: oracleHelmOfObedience,
	})
	advanceTo(t, g, game.StepPrecombatMain)

	// No mana at all. "X can't be 0", so there is no legal
	// activation — NOT a free one at X=0. Offering it would be #544:
	// an enumeration the engine rejects.
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), helm); len(acts) != 0 {
		t.Fatalf("with no mana the Helm has no legal activation, got %d: %v", len(acts), labels(acts))
	}

	mana(g, active, 2)
	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, helm)
	if len(acts) == 0 {
		t.Fatal("two mana should buy an activation")
	}
	for _, m := range acts {
		if got := xValueOf(t, m); got != 2 {
			t.Errorf("X = %d, want 2", got)
		}
	}
	dispatchAll(t, g, active.ID, moves)
}

func TestXOnAnAbilityDoesNotEatTheTargetExpansionBudget(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	helm := battlefieldCard(g, active, game.Card{
		Name: "Helm of Obedience", TypeLine: "Artifact", ManaCost: "{4}", OracleID: oracleHelmOfObedience,
	})
	mana(g, active, 6)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, helm)
	// Three opponents at a four-seat table. X collapses to ONE
	// value, so the budget buys one activation per opponent — which
	// is the whole design point: an X that enumerated 1..MaxX would
	// spend the twelve-move budget on twelve near-identical
	// activations pointed at the first opponent and never offer the
	// other two. That is #544's shape exactly.
	if len(acts) != 3 {
		t.Fatalf("want one activation per opponent, got %d: %v", len(acts), labels(acts))
	}
	seen := map[string]bool{}
	for _, m := range acts {
		var p struct {
			Targets []struct {
				ID string `json:"id"`
			} `json:"targets"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("params: %v", err)
		}
		if len(p.Targets) != 1 {
			t.Fatalf("expected one target per move: %s", string(m.Params))
		}
		seen[p.Targets[0].ID] = true
		if got := xValueOf(t, m); got != 6 {
			t.Errorf("X = %d, want 6 — every opponent gets the same affordable X", got)
		}
	}
	if len(seen) != 3 {
		t.Errorf("three distinct opponents should be offered, got %d", len(seen))
	}
	dispatchAll(t, g, active.ID, moves)
}

func TestSoothsayingOffersBothAbilitiesIndependently(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	sooth := battlefieldCard(g, active, game.Card{
		Name: "Soothsaying", TypeLine: "Enchantment", ManaCost: "{U}", OracleID: oracleSoothsaying,
	})
	// Three colorless: enough for the {X} at X=3, not enough for the
	// {3}{U}{U} shuffle, which needs two blue.
	mana(g, active, 3)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, sooth)
	if len(acts) != 1 {
		t.Fatalf("only the {X} half is affordable, got %d: %v", len(acts), labels(acts))
	}
	if got := xValueOf(t, acts[0]); got != 3 {
		t.Errorf("X = %d, want 3", got)
	}
	dispatchAll(t, g, active.ID, moves)
}

func TestCrewIsOfferedWithACrewSetTheEngineAccepts(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	copter := battlefieldCard(g, active, game.Card{
		Name: "Smuggler's Copter", TypeLine: "Artifact — Vehicle", ManaCost: "{2}",
		Power: 3, Toughness: 3, OracleID: oracleSmugglersCopter,
	})
	advanceTo(t, g, game.StepPrecombatMain)

	// Nothing to crew with: the ability is not offered at all,
	// rather than offered with an empty crew_ids the engine bounces.
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), copter); len(acts) != 0 {
		t.Fatalf("no creatures → no crew activation, got %v", labels(acts))
	}

	battlefieldCard(g, active, creature("Crewer", "{1}{G}", 2, 2))
	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, copter)
	if len(acts) != 1 {
		t.Fatalf("want one crew activation, got %d: %v", len(acts), labels(acts))
	}
	var p struct {
		CrewIDs []string `json:"crew_ids"`
	}
	if err := json.Unmarshal(acts[0].Params, &p); err != nil {
		t.Fatalf("params: %v", err)
	}
	if len(p.CrewIDs) == 0 {
		t.Error("a crew activation must name the creatures that pay for it")
	}
	// The whole reason this test exists: before the crew payment was
	// enumerated, this move was offered and the engine refused it.
	dispatchAll(t, g, active.ID, moves)
}
