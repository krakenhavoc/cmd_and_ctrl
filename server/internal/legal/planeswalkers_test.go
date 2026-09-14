package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// planeswalkers_test.go — the enumerator half of S27's planeswalkers.
//
// The invariant is #544's, and a hung table is what breaks when it
// fails: a bot seat that is offered a loyalty activation the engine
// then refuses has no move to make and the table stops. Loyalty
// abilities are a fresh cost component with four gates of their own
// (CR 606.1 / 606.2 / 606.3 / 606.5), so each new card is a fresh
// chance for the enumerator and ActivateCatalogAbility to disagree
// about whether an activation is legal.
//
// dispatchAll is the assertion that matters. Everything above it is
// setup to make sure there is something to dispatch.

const (
	oracleUgin          = "eecb3047-a563-441a-9175-200421981ac3"
	oracleLiliana       = "0ba134d8-ee7d-48ec-8dc6-57942b8e9261"
	oracleTeferiPilgrim = "5f0fa785-37ab-46a8-9ec0-6f8b29576d93"
	oracleTeferiHero    = "f2f165b6-ef0a-42ad-9352-ba68be8248b0"
	oracleKarn          = "0ca233f4-1b7f-4807-ab6e-2b1f5439b3db"
	oracleSaheeli       = "5f3fe679-aff1-41b3-8d75-c78c2c0636f0"
)

// walker puts a planeswalker on the battlefield with enough loyalty
// that every one of its abilities is affordable, so the enumerator
// has to make a judgement about all of them rather than being let
// off by CR 606.3.
func walker(g *game.Game, p *game.Player, name, oracleID string, loyalty int) uuid.UUID {
	return battlefieldCard(g, p, game.Card{
		Name:     name,
		OracleID: oracleID,
		TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
}

// TestS27PlaneswalkerLoyaltyAbilitiesAreOfferedAndAccepted is the
// #544 guard for all six cards at once: every loyalty activation the
// enumerator advertises must be one the dispatcher takes.
func TestS27PlaneswalkerLoyaltyAbilitiesAreOfferedAndAccepted(t *testing.T) {
	for _, tc := range []struct {
		name    string
		oracle  string
		loyalty int
		want    int // abilities the catalog registers
	}{
		{"Ugin, the Spirit Dragon", oracleUgin, 14, 2},
		{"Liliana of the Veil", oracleLiliana, 8, 2},
		{"Teferi, Temporal Pilgrim", oracleTeferiPilgrim, 8, 2},
		{"Teferi, Hero of Dominaria", oracleTeferiHero, 8, 2},
		{"Karn Liberated", oracleKarn, 8, 2},
		{"Saheeli Rai", oracleSaheeli, 10, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newTable(t)
			active := g.Seats[g.Turn.ActiveSeat]
			clearHand(active)
			// Something to point the targeted abilities at, on both
			// sides: Saheeli's −2 copies a permanent YOU control,
			// and with no such permanent the enumerator is right to
			// leave the ability out.
			battlefieldCard(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)],
				creature("Their Bear", "{1}{G}", 2, 2))
			battlefieldCard(g, active, creature("Your Bear", "{1}{G}", 2, 2))
			pw := walker(g, active, tc.name, tc.oracle, tc.loyalty)
			advanceTo(t, g, game.StepPrecombatMain)

			moves := legal.EnumerateFor(g, active.ID)
			acts := activationsOf(moves, pw)
			if len(acts) < tc.want {
				t.Fatalf("%d activations offered, want at least %d (one per registered ability): %v",
					len(acts), tc.want, labels(acts))
			}
			for _, m := range acts {
				if m.Cost == nil || m.Cost.Loyalty == 0 {
					// Teferi's [0] really is a zero cost; the
					// others must carry theirs so a policy can
					// tell a plus from a minus.
					if tc.oracle != oracleTeferiPilgrim {
						t.Errorf("activation %q carries no loyalty cost", m.Label)
					}
				}
			}
			// The whole point: nothing offered may be refused.
			dispatchAll(t, g, active.ID, moves)
		})
	}
}

// TestLoyaltyActivationIsNotOfferedTwiceInATurn mirrors
// ActivateCatalogAbility's CR 606.5 gate. Offering a second
// activation the engine refuses is the hung-table shape.
func TestLoyaltyActivationIsNotOfferedTwiceInATurn(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	pw := walker(g, active, "Karn Liberated", oracleKarn, 6)
	advanceTo(t, g, game.StepPrecombatMain)

	if n := len(activationsOf(legal.EnumerateFor(g, active.ID), pw)); n == 0 {
		t.Fatal("no loyalty activation offered on an untouched planeswalker")
	}
	// Ability 1 is the −3; it needs no mana and no sacrifice.
	err := g.ActivateCatalogAbility(active.ID, pw, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: pw}},
	})
	if err != nil {
		t.Fatalf("−3: %v", err)
	}
	if n := len(activationsOf(legal.EnumerateFor(g, active.ID), pw)); n != 0 {
		t.Errorf("%d activations still offered after the turn's loyalty ability", n)
	}
}

// TestUnaffordableMinusIsNotOffered is CR 606.3 from the
// enumerator's side: a walker at 1 loyalty cannot pay Karn's −3, and
// a move list that offered it would be advertising a refusal.
func TestUnaffordableMinusIsNotOffered(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	pw := walker(g, active, "Karn Liberated", oracleKarn, 1)
	battlefieldCard(g, active, creature("Bear", "{1}{G}", 2, 2))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	for _, m := range activationsOf(moves, pw) {
		if m.Cost != nil && m.Cost.Loyalty < 0 {
			t.Errorf("offered %q with only 1 loyalty on the card", m.Label)
		}
	}
	dispatchAll(t, g, active.ID, moves)
}

// TestLoyaltyAbilitiesAreNotOfferedAtInstantSpeed is CR 606.5's
// other half.
func TestLoyaltyAbilitiesAreNotOfferedAtInstantSpeed(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	pw := walker(g, active, "Saheeli Rai", oracleSaheeli, 10)
	advanceTo(t, g, game.StepDeclareAttackers)

	if n := len(activationsOf(legal.EnumerateFor(g, active.ID), pw)); n != 0 {
		t.Errorf("%d loyalty activations offered outside a main phase", n)
	}
}
