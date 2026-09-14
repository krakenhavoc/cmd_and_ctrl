package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// ability_removal_test.go states the invariant this package exists to
// keep, for CR 613.1f: the enumerator must not offer an ability the
// permanent no longer has.
//
// It is not a cosmetic concern. #544: a seat that owes a decision is
// enumerated that decision's answers and nothing else, so a move list
// the engine refuses is not a wasted turn, it is a hung table — a bot
// retries the same rejected move forever and holds every other seat
// with it. The enumerator agreeing with the engine is the whole
// contract, and the reason `game.ActivatedAbilitiesForCard` /
// `game.ManaAbilitiesForCard` are the seam rather than a check
// written here: two copies of a rule drift, one does not.

const (
	oracleSongOfTheDryads = "7c3944fa-7c86-4979-85a9-86196aa94594"
	oracleSolRingLegal    = "6ad8011d-3471-4369-9d68-b264cc027487"
	oracleBonesplitter    = "452e3f5f-ce17-4682-966b-5cc100210aee"
)

// songOnto puts a Song of the Dryads on the battlefield already
// attached to `host`, and invalidates the layer cache so the next
// read recomputes. Hand-built rather than cast, because what is
// under test is the enumerator, not the cast path.
func songOnto(t *testing.T, g *game.Game, p *game.Player, host uuid.UUID) {
	t.Helper()
	battlefieldCard(g, p, game.Card{
		Name: "Song of the Dryads", TypeLine: "Enchantment — Aura",
		OracleID:   oracleSongOfTheDryads,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
	g.BumpLayerVersionForTest()
}

func movesOfKindFor(moves []legal.Move, kind legal.Kind, source uuid.UUID) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Kind == kind && m.Source == source {
			out = append(out, m)
		}
	}
	return out
}

// An Equipment whose abilities were removed offers no equip. Before
// the layer-6 seam the enumerator read the catalog directly and would
// have offered it; the engine's ActivateCatalogAbility reads the same
// accessor and would have refused it.
func TestEnumeratorDropsAnEquipAbilityAfterAbilityRemoval(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	splitter := battlefieldCard(g, active, game.Card{
		Name: "Bonesplitter", TypeLine: "Artifact — Equipment",
		OracleID: oracleBonesplitter,
	})
	battlefieldCard(g, active, creature("Bear", "{1}{G}", 2, 2))
	mana(g, active, 3)
	advanceTo(t, g, game.StepPrecombatMain)

	before := movesOfKindFor(legal.EnumerateFor(g, active.ID), legal.KindActivate, splitter)
	if len(before) == 0 {
		t.Fatal("fixture is wrong: the equip should be enumerable before the Song lands")
	}

	songOnto(t, g, active, splitter)

	if got := movesOfKindFor(legal.EnumerateFor(g, active.ID), legal.KindActivate, splitter); len(got) != 0 {
		t.Errorf("enumerated %d activated abilities on a permanent with none: %+v", len(got), got)
	}
}

// The mana half, and the CR 305.7 half of it: the declared "{T}: Add
// {C}{C}" stops being enumerated and the Forest's intrinsic "{T}: Add
// {G}" starts. One move either way, so a count assertion alone would
// have passed the wrong answer.
func TestEnumeratorSwapsADeclaredManaAbilityForTheLandTypesOne(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	solRing := battlefieldCard(g, active, game.Card{
		Name: "Sol Ring", TypeLine: "Artifact", OracleID: oracleSolRingLegal,
	})
	advanceTo(t, g, game.StepPrecombatMain)

	before := movesOfKindFor(legal.EnumerateFor(g, active.ID), legal.KindMana, solRing)
	if len(before) != 1 {
		t.Fatalf("fixture is wrong: Sol Ring should offer one mana move, got %+v", before)
	}

	songOnto(t, g, active, solRing)

	after := movesOfKindFor(legal.EnumerateFor(g, active.ID), legal.KindMana, solRing)
	if len(after) != 1 {
		t.Fatalf("mana moves = %+v, want exactly the Forest's one", after)
	}
	if before[0].Label == after[0].Label {
		t.Errorf("the offered ability did not change: still %q", after[0].Label)
	}
}

// The end-to-end version of the #544 invariant for this seam: every
// move the enumerator offers on a silenced permanent is one the
// ENGINE accepts. Both halves are asserted from the same board,
// because "offer nothing" and "offer only what works" are different
// promises and only the second one is true here — a Song of the
// Dryads leaves its host a Forest, and a Forest taps for {G}.
func TestEveryEnumeratedMoveIsStillAcceptedUnderAbilityRemoval(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	splitter := battlefieldCard(g, active, game.Card{
		Name: "Bonesplitter", TypeLine: "Artifact — Equipment",
		OracleID: oracleBonesplitter,
	})
	battlefieldCard(g, active, creature("Bear", "{1}{G}", 2, 2))
	mana(g, active, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	songOnto(t, g, active, splitter)

	var offered int
	for _, m := range legal.EnumerateFor(g, active.ID) {
		if m.Source != splitter {
			continue
		}
		offered++
		if m.Kind != legal.KindMana {
			t.Fatalf("a silenced permanent offered a non-mana move: %+v", m)
		}
		if err := g.ActivateManaAbility(active.ID, splitter, 0, game.ManaAbilityParams{}); err != nil {
			t.Errorf("the engine refused an enumerated move (%q): %v", m.Label, err)
		}
	}
	if offered != 1 {
		t.Errorf("offered %d moves on the silenced Equipment, want just the Forest's mana ability", offered)
	}
}
