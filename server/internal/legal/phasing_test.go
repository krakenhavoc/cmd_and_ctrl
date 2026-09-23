package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// phasing_test.go — #1199, ADR 0084, the enumerator's half.
//
// The claim ADR 0084 Decision 1 makes is that the enumerator needs NO
// CODE for phasing: it walks g.Battlefield.Cards and reads
// legalTargetsLocked, and a phased-out permanent is in neither. This
// file is that claim asserted rather than assumed — the #499/#618
// agreement rule, in the direction that matters for a bot. A bot
// offered a move on a permanent that does not exist picks it, is
// refused, and picks it again.

// oracleVodalianIllusionist is the proof card's key: "{U}{U}, {T}:
// Target creature phases out."
const (
	oracleVodalianIllusionist = "ac935639-ba30-4b04-86f6-363527e85a8b"
	oracleRealityRipple       = "504d5c29-7c37-4c31-8549-2a49eeef74c8"
)

// A phased-out permanent's own activated ability is not a move.
func TestAPhasedOutPermanentsAbilityIsNotEnumerated(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	illusionist := battlefieldCard(g, active, game.Card{
		Name: "Vodalian Illusionist", TypeLine: "Creature — Merfolk Wizard",
		OracleID: oracleVodalianIllusionist, Power: 1, Toughness: 2,
	})
	battlefieldCard(g, active, game.Card{
		Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
	})
	for range 4 {
		battlefieldCard(g, active, basic("Island", "Island"))
	}
	advanceTo(t, g, game.StepPrecombatMain)

	// Without the phase-out it is a move, or the absence below would
	// prove nothing.
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), illusionist); len(acts) == 0 {
		t.Fatal("the Illusionist's ability should be a move with nothing restricting it")
	}

	g.WithWriteLock(func() {
		if err := g.PhaseOutForEffect(uuid.Nil, illusionist); err != nil {
			t.Fatalf("phase out: %v", err)
		}
	})

	if acts := activationsOf(legal.EnumerateFor(g, active.ID), illusionist); len(acts) != 0 {
		t.Errorf("a phased-out permanent's ability is still offered (CR 702.26b): %v", labels(acts))
	}
}

// A phased-out permanent is not a target the enumerator will offer,
// which is the same fact the client's `legal_targets` reads.
func TestAPhasedOutPermanentIsNotAnEnumeratedTarget(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	victim := battlefieldCard(g, active, game.Card{
		Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
	})
	stays := battlefieldCard(g, active, game.Card{
		Name: "Present Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
	})
	for range 4 {
		battlefieldCard(g, active, basic("Island", "Island"))
	}
	advanceTo(t, g, game.StepPrecombatMain)

	// Reality Ripple's clause is on the CARD, so this is the same
	// TargetSpec the client's `legal_targets` and the bot's cast
	// enumeration both read.
	spec := game.TargetSpecFor(oracleRealityRipple)
	if spec == nil {
		t.Fatal("Reality Ripple declares a card-level target clause")
	}

	g.WithWriteLock(func() {
		if err := g.PhaseOutForEffect(uuid.Nil, victim); err != nil {
			t.Fatalf("phase out: %v", err)
		}
	})

	g.WithWriteLock(func() {
		lt := g.LegalTargetsForEffect(game.TargetSource{Controller: active.ID}, spec)
		sawStays := false
		for _, id := range lt.Cards {
			if id == victim {
				t.Error("a phased-out creature is still offered as a target (CR 702.26b)")
			}
			if id == stays {
				sawStays = true
			}
		}
		if !sawStays {
			t.Error("the creature that did not phase out is still a target")
		}
	})
}
