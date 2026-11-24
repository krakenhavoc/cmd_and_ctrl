package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b11_cards_test.go — GAP2 batch 11 (#5948) on seam #6098: each
// test plays the real catalog entry through combat and asserts what its
// text says.

const (
	gap2B11ThijarianOracle = "d273033b-3b8b-4ce9-afe8-208d82dd20f1"
	gap2B11FirkraagOracle  = "81dad7c5-5950-4c45-adb0-cf1a15a74250"
)

func gap2B11Registered(t *testing.T, oracle, name string) {
	t.Helper()
	spec, ok := effects.Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", name)
	}
	if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
	}
}

func gap2B11Clues(g *game.Game, who uuid.UUID) int {
	return effectstest.TokensNamedControlledBy(g, "Clue", who)
}

// --- Thijarian Witness -------------------------------------------------------

// A lone attacker that dies to its lone blocker was attacking alone: it is
// exiled and the Witness's controller investigates.
func TestGap2B11ThijarianWitnessExilesALoneAttacker(t *testing.T) {
	gap2B11Registered(t, gap2B11ThijarianOracle, "Thijarian Witness")
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, opp.ID, "Thijarian Witness", "Creature — Alien Cleric", gap2B11ThijarianOracle, 0, 4)
	att := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blk := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.Pr7bAttack(t, g, opp.ID, att)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blk: att})
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, att) {
		t.Error("the lone attacker that died is not in exile")
	}
	if n := gap2B11Clues(g, opp.ID); n != 1 {
		t.Errorf("the Witness's controller has %d Clues, want 1", n)
	}
}

// A lone blocker that dies was blocking alone.
func TestGap2B11ThijarianWitnessExilesALoneBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Thijarian Witness", "Creature — Alien Cleric", gap2B11ThijarianOracle, 0, 4)
	att := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	blk := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.Pr7bAttack(t, g, opp.ID, att)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blk: att})
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, blk) {
		t.Error("the lone blocker that died is not in exile")
	}
	if n := gap2B11Clues(g, me.ID); n != 1 {
		t.Errorf("%d Clues, want 1", n)
	}
}

// Two attackers that die together in one combat damage step were not
// attacking alone, though one of them leaves the battlefield after the
// other: no trigger, nothing exiled.
func TestGap2B11ThijarianWitnessIgnoresAttackersThatDieTogether(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, opp.ID, "Thijarian Witness", "Creature — Alien Cleric", gap2B11ThijarianOracle, 0, 4)
	a1 := effectstest.B12Creature(g, me.ID, "Bear One", "Creature — Bear", 2, 2)
	a2 := effectstest.B12Creature(g, me.ID, "Bear Two", "Creature — Bear", 2, 2)
	b1 := effectstest.B12Creature(g, opp.ID, "Ogre One", "Creature — Ogre", 3, 3)
	b2 := effectstest.B12Creature(g, opp.ID, "Ogre Two", "Creature — Ogre", 3, 3)
	effectstest.Pr7bAttack(t, g, opp.ID, a1, a2)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{b1: a1, b2: a2})
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a1, a2} {
		if effectstest.FindBattlefieldCardForTest(g, id) != nil {
			t.Fatalf("setup: attacker %s survived", id)
		}
		if effectstest.InExile(g, id) {
			t.Errorf("attacker %s was exiled, but it was not attacking alone", id)
		}
	}
	if n := gap2B11Clues(g, opp.ID); n != 0 {
		t.Errorf("%d Clues, want none", n)
	}
}

// A creature that dies outside combat, and the Witness itself, do not
// trigger it.
func TestGap2B11ThijarianWitnessIgnoresACreatureOutOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Thijarian Witness", "Creature — Alien Cleric", gap2B11ThijarianOracle, 0, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.InExile(g, bear) || gap2B11Clues(g, me.ID) != 0 {
		t.Error("a creature out of combat triggered the Witness")
	}
}

// --- Firkraag, Cunning Instigator -------------------------------------------

// A Dragon attacking an opponent goads a creature that player controls.
func TestGap2B11FirkraagGoadsACreatureOfTheAttackedOpponent(t *testing.T) {
	gap2B11Registered(t, gap2B11FirkraagOracle, "Firkraag, Cunning Instigator")
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	firk := effectstest.B12Push(g, me.ID, "Firkraag, Cunning Instigator", "Legendary Creature — Dragon", gap2B11FirkraagOracle, 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	untouched := effectstest.B12Creature(g, other.ID, "Third Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(firk, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B33Goaded(t, g, theirs) != me.ID {
		t.Error("the chosen creature is not goaded by me")
	}
	if effectstest.B33Goaded(t, g, untouched) != uuid.Nil {
		t.Error("a creature of a player I did not attack was goaded")
	}
}

// A non-Dragon attacking does not trigger the goad.
func TestGap2B11FirkraagIgnoresANonDragonAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Firkraag, Cunning Instigator", "Legendary Creature — Dragon", gap2B11FirkraagOracle, 3, 3)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a non-Dragon attack asked for a goad target")
	}
}

// A creature that had to attack (goaded as attackers were declared) deals
// combat damage to an opponent: a +1/+1 counter on Firkraag and a card. A
// creature that did not have to attack does nothing.
func TestGap2B11FirkraagRewardsACreatureThatHadToAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	firk := effectstest.B12Push(g, me.ID, "Firkraag, Cunning Instigator", "Legendary Creature — Dragon", gap2B11FirkraagOracle, 3, 3)
	goaded := effectstest.B12Creature(g, me.ID, "Goaded Bear", "Creature — Bear", 2, 2)
	free := effectstest.B12Creature(g, me.ID, "Free Bear", "Creature — Bear", 2, 2)
	if err := g.SetGoaded(goaded, opp.ID); err != nil {
		t.Fatal(err)
	}
	hand := me.Hand.Size()
	effectstest.Pr7bAttack(t, g, other.ID, goaded, free)
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, firk, "+1/+1"); n != 1 {
		t.Errorf("Firkraag has %d +1/+1 counters, want 1 (only the goaded creature had to attack)", n)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

// A creature goaded only after attackers were declared never had to
// attack: its combat damage does not trigger Firkraag.
func TestGap2B11FirkraagIgnoresACreatureGoadedAfterItAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	firk := effectstest.B12Push(g, me.ID, "Firkraag, Cunning Instigator", "Legendary Creature — Dragon", gap2B11FirkraagOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.Pr7bAttack(t, g, other.ID, bear)
	if err := g.SetGoaded(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	hand := me.Hand.Size()
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, firk, "+1/+1"); n != 0 {
		t.Errorf("Firkraag has %d +1/+1 counters, want 0", n)
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d", got, hand)
	}
}
