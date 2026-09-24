package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// activation_timing_test.go — the enumerator half of #1208, in the
// #499/#618 agreement style: every activation the enumerator OFFERS
// is one the engine accepts, and every activation the engine would
// REFUSE for timing is one the enumerator does not offer.
//
// That is the whole reason the two read one function. Before #1208
// the enumerator kept its own copy of `(ab.SorcerySpeed ||
// ab.Cost.Loyalty != nil) && !speed`, so a bot under a Leonin Shikari
// would never have been offered an equip outside its main phase — a
// move the engine accepts, withheld, which is #544's defect with the
// sign flipped.

// oracleBonesplitter is ability_removal_test.go's, reused rather
// than respelled — one Equipment with one equip ability is all this
// file needs.
const (
	oracleLeoninShikari     = "857d94f7-113c-45a4-a88a-5d087347d57f"
	oracleTimeRavelerTiming = "ae7604bb-4818-45a3-960c-cf3d83f15964"
)

// TestLeoninShikariPutsEquipOnTheListOutsideTheSorceryWindow is the
// grant's #544 half: an equip ability is sorcery-speed on its own
// (CR 702.6a), and the Shikari's static opens it.
func TestLeoninShikariPutsEquipOnTheListOutsideTheSorceryWindow(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	battlefieldCard(g, active, game.Card{
		Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2,
	})
	// Equip {1} has to be affordable, or the absence below would be
	// about the mana rather than about the timing.
	lands(g, active, "Plains", "Plains", 2)
	sword := battlefieldCard(g, active, game.Card{
		Name: "Bonesplitter", TypeLine: "Artifact — Equipment",
		OracleID: oracleBonesplitter,
	})
	// The active seat's END step: it holds priority, and the sorcery
	// window (CR 307.1) is shut because the step is not a main phase.
	advanceTo(t, g, game.StepEnd)

	moves := legal.EnumerateFor(g, active.ID)
	if acts := activationsOf(moves, sword); len(acts) != 0 {
		t.Fatalf("setup: equip offered in an end step with nothing granting: %v", labels(acts))
	}

	battlefieldCard(g, active, game.Card{
		Name: "Leonin Shikari", TypeLine: "Creature — Cat Soldier",
		OracleID: oracleLeoninShikari, Power: 2, Toughness: 2,
	})

	moves = legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, sword)
	if len(acts) == 0 {
		t.Fatal("under a Leonin Shikari, equip is still withheld outside the sorcery window")
	}
	// #544 the other way: what the list offers, the engine accepts.
	dispatchAll(t, g, active.ID, acts)
}

// The Shikari's clause names EQUIP abilities, so a loyalty ability
// on the same board stays sorcery-speed — the enumerator's half of
// the narrowing `ActivationAbility.Equip` exists for.
func TestLeoninShikariDoesNotOpenALoyaltyAbility(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	lands(g, active, "Plains", "Plains", 2)
	teferi := battlefieldCard(g, active, game.Card{
		Name: "Teferi, Time Raveler", TypeLine: "Legendary Planeswalker — Teferi",
		OracleID: oracleTimeRavelerTiming, Counters: map[string]int{"loyalty": 4},
	})
	battlefieldCard(g, active, game.Card{
		Name: "Leonin Shikari", TypeLine: "Creature — Cat Soldier",
		OracleID: oracleLeoninShikari, Power: 2, Toughness: 2,
	})

	// Main phase: the loyalty rows ARE moves, so their absence in the
	// end step below is about the window and not about the fixture.
	advanceTo(t, g, game.StepPrecombatMain)
	if len(activationsOf(legal.EnumerateFor(g, active.ID), teferi)) == 0 {
		t.Fatal("setup: Teferi's loyalty abilities are not moves in a main phase")
	}

	advanceTo(t, g, game.StepEnd)
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), teferi); len(acts) != 0 {
		t.Errorf("the Shikari opened a loyalty ability it does not name: %v", labels(acts))
	}
}
