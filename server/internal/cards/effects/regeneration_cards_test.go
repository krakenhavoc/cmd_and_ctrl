package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// regeneration_cards_test.go is the catalog half of #667 (CR 701.19).
// The engine half — what a shield is, what spends one, and what
// ignores one — is pinned in game/regeneration_test.go. What is here
// is the four cards that make shields and the family of cards that
// print "can't be regenerated".

const (
	asceticismOracle       = "9396546c-d067-4ce3-9c3d-b62cca970b4f"
	wrapInVigorOracle      = "39da2aa8-f4d9-44f6-a446-488beaec821f"
	weldingJarOracle       = "e4a6c421-bef1-4360-a2dc-ae9c61d86b2f"
	goblinChirurgeonOracle = "ea55db87-a5c3-49f7-b968-77510cdeb469"
	damnationOracle        = "d57a8f0b-7989-4db5-8756-6f2690097252"
	dayOfJudgmentOracle    = "d057289d-5e28-43d5-8ff3-4a1bc723477d"
)

// regenShieldsOn reads the shield count off a battlefield permanent.
func regenShieldsOn(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return c.RegenerationShields
}

// destroyIt is a bare "destroy target permanent" from outside the
// catalog, standing in for whatever removal the opponent has.
func destroyIt(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
}

// --- Asceticism ---------------------------------------------------

// The card #667 exists for. Both halves, on one board: the creatures
// have hexproof, and {1}{G} puts a shield on one of them that
// replaces the next destruction.
func TestAsceticismGrantsHexproofAndRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedCreature(g, "Bear", me.ID)
	asceticism := pushCatalogPermanent(g, me.ID, "Asceticism", "Enchantment", asceticismOracle, false)

	if !containsString(effectiveAbilities(t, g, bear), "hexproof") {
		t.Error("Creatures you control have hexproof")
	}

	if err := g.ActivateCatalogAbility(me.ID, asceticism, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if n := regenShieldsOn(t, g, bear); n != 1 {
		t.Fatalf("shields = %d, want 1 after the ability resolved", n)
	}

	destroyIt(t, g, bear)

	c, ok := battlefieldCard(g, bear)
	if !ok {
		t.Fatal("CR 701.19a: the shield replaces the destruction, so the creature is still there")
	}
	if !c.Tapped {
		t.Error("a regenerated creature is tapped")
	}
	if c.RegenerationShields != 0 {
		t.Errorf("shields = %d, want 0 — the shield was used up", c.RegenerationShields)
	}
}

// Repeatable, and the shields stack — which is the reason the card is
// worth five mana.
func TestAsceticismCanShieldTwice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedCreature(g, "Bear", me.ID)
	asceticism := pushCatalogPermanent(g, me.ID, "Asceticism", "Enchantment", asceticismOracle, false)

	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, asceticism, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
	}
	if n := regenShieldsOn(t, g, bear); n != 2 {
		t.Fatalf("shields = %d, want 2", n)
	}

	destroyIt(t, g, bear)
	destroyIt(t, g, bear)
	if _, ok := battlefieldCard(g, bear); !ok {
		t.Error("two shields survive two destructions")
	}
}

// --- Wrap in Vigor ------------------------------------------------

// "Regenerate each creature you control" — yours, and only yours.
func TestWrapInVigorShieldsYourCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	mine := seedCreature(g, "Bear", me.ID)
	alsoMine := seedCreature(g, "Elk", me.ID)
	theirs := seedCreature(g, "Ogre", them.ID)

	castCatalogSpell(t, g, "Wrap in Vigor", "Instant", wrapInVigorOracle, nil)
	passPriorityAroundTable(t, g)

	if n := regenShieldsOn(t, g, mine); n != 1 {
		t.Errorf("my first creature has %d shields, want 1", n)
	}
	if n := regenShieldsOn(t, g, alsoMine); n != 1 {
		t.Errorf("my second creature has %d shields, want 1", n)
	}
	if n := regenShieldsOn(t, g, theirs); n != 0 {
		t.Errorf("an opponent's creature has %d shields, want 0 — \"each creature YOU control\"", n)
	}

	destroyIt(t, g, mine)
	destroyIt(t, g, theirs)
	if _, ok := battlefieldCard(g, mine); !ok {
		t.Error("my creature should have regenerated")
	}
	if _, ok := battlefieldCard(g, theirs); ok {
		t.Error("their creature had no shield")
	}
}

// --- Welding Jar --------------------------------------------------

// The Jar eats itself to shield an artifact. The sacrifice is a cost,
// so it is gone before the ability resolves.
func TestWeldingJarSacrificesItselfToShieldAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rock := pushCatalogPermanent(g, me.ID, "Mana Rock", "Artifact", "test-regen-rock", false)
	jar := pushCatalogPermanent(g, me.ID, "Welding Jar", "Artifact", weldingJarOracle, false)

	if err := g.ActivateCatalogAbility(me.ID, jar, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, ok := battlefieldCard(g, jar); ok {
		t.Error("the sacrifice is a cost, paid at announce")
	}
	passPriorityAroundTable(t, g)

	if n := regenShieldsOn(t, g, rock); n != 1 {
		t.Fatalf("shields on the artifact = %d, want 1", n)
	}
	destroyIt(t, g, rock)
	c, ok := battlefieldCard(g, rock)
	if !ok {
		t.Fatal("the artifact should have regenerated")
	}
	if !c.Tapped {
		t.Error("a regenerated artifact is tapped too — CR 701.19a says permanent, not creature")
	}
}

// --- Goblin Chirurgeon --------------------------------------------

// "Sacrifice a Goblin" — another Goblin, here, so the Chirurgeon is
// still around to be read.
func TestGoblinChirurgeonEatsAGoblinToRegenerate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	chirurgeon := pushCatalogPermanent(g, me.ID, "Goblin Chirurgeon", "Creature — Goblin Shaman", goblinChirurgeonOracle, false)
	fodder := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Token", TypeLine: "Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := seedCreature(g, "Bear", me.ID)

	if err := g.ActivateCatalogAbility(me.ID, chirurgeon, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, ok := battlefieldCard(g, fodder); ok {
		t.Error("the sacrificed Goblin is a cost, paid at announce")
	}
	passPriorityAroundTable(t, g)

	if n := regenShieldsOn(t, g, bear); n != 1 {
		t.Fatalf("shields = %d, want 1", n)
	}
	destroyIt(t, g, bear)
	if _, ok := battlefieldCard(g, bear); !ok {
		t.Error("the shield should have replaced the destruction")
	}
}

// --- "can't be regenerated" ---------------------------------------

// The clause that was cosmetic on every card that printed it until
// there was a shield to ignore. Damnation kills through one;
// Day of Judgment, which does NOT print the clause, does not.
func TestDamnationKillsThroughAShieldAndDayOfJudgmentDoesNot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	// Day of Judgment first: the shield holds.
	survivor := seedCreature(g, "Bear", me.ID)
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(survivor); err != nil {
			t.Fatalf("RegenerateForEffect: %v", err)
		}
	})
	castCatalogSpell(t, g, "Day of Judgment", "Sorcery", dayOfJudgmentOracle, nil)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, survivor); !ok {
		t.Fatal("Day of Judgment does not say they can't be regenerated, so the shield replaces its destruction")
	}

	// Damnation next, on a fresh shield: it does not.
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(survivor); err != nil {
			t.Fatalf("RegenerateForEffect: %v", err)
		}
	})
	castCatalogSpell(t, g, "Damnation", "Sorcery", damnationOracle, nil)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, survivor); ok {
		t.Error("CR 701.19c: \"They can't be regenerated\" ignores the shield")
	}
}

// The single-target family carries the same rider — Terminate is the
// one that shares its whole body with Mortify and Putrefy.
func TestTerminateIgnoresARegenerationShield(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	theirs := seedCreature(g, "Bear", them.ID)
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(theirs); err != nil {
			t.Fatalf("RegenerateForEffect: %v", err)
		}
	})
	_ = me

	castCatalogSpell(t, g, "Terminate", "Instant", terminateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)

	if _, ok := battlefieldCard(g, theirs); ok {
		t.Error("CR 701.19c: Terminate says it can't be regenerated")
	}
	if !them.Graveyard.Contains(theirs) {
		t.Error("the creature should be in its owner's graveyard")
	}
}
