package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Asceticism ---------------------------------------------------

// The card #667 exists for. Both halves, on one board: the creatures
// have hexproof, and {1}{G} puts a shield on one of them that
// replaces the next destruction.
func TestAsceticismGrantsHexproofAndRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	asceticism := effectstest.PushCatalogPermanent(g, me.ID, "Asceticism", "Enchantment", effectstest.AsceticismOracle, false)

	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "hexproof") {
		t.Error("Creatures you control have hexproof")
	}

	if err := g.ActivateCatalogAbility(me.ID, asceticism, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.RegenShieldsOn(t, g, bear); n != 1 {
		t.Fatalf("shields = %d, want 1 after the ability resolved", n)
	}

	effectstest.DestroyIt(t, g, bear)

	c, ok := effectstest.BattlefieldCard(g, bear)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	asceticism := effectstest.PushCatalogPermanent(g, me.ID, "Asceticism", "Enchantment", effectstest.AsceticismOracle, false)

	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, asceticism, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.RegenShieldsOn(t, g, bear); n != 2 {
		t.Fatalf("shields = %d, want 2", n)
	}

	effectstest.DestroyIt(t, g, bear)
	effectstest.DestroyIt(t, g, bear)
	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Error("two shields survive two destructions")
	}
}

// --- Wrap in Vigor ------------------------------------------------

// "Regenerate each creature you control" — yours, and only yours.
func TestWrapInVigorShieldsYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	mine := effectstest.SeedCreature(g, "Bear", me.ID)
	alsoMine := effectstest.SeedCreature(g, "Elk", me.ID)
	theirs := effectstest.SeedCreature(g, "Ogre", them.ID)

	effectstest.CastCatalogSpell(t, g, "Wrap in Vigor", "Instant", effectstest.WrapInVigorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.RegenShieldsOn(t, g, mine); n != 1 {
		t.Errorf("my first creature has %d shields, want 1", n)
	}
	if n := effectstest.RegenShieldsOn(t, g, alsoMine); n != 1 {
		t.Errorf("my second creature has %d shields, want 1", n)
	}
	if n := effectstest.RegenShieldsOn(t, g, theirs); n != 0 {
		t.Errorf("an opponent's creature has %d shields, want 0 — \"each creature YOU control\"", n)
	}

	effectstest.DestroyIt(t, g, mine)
	effectstest.DestroyIt(t, g, theirs)
	if _, ok := effectstest.BattlefieldCard(g, mine); !ok {
		t.Error("my creature should have regenerated")
	}
	if _, ok := effectstest.BattlefieldCard(g, theirs); ok {
		t.Error("their creature had no shield")
	}
}

// --- Welding Jar --------------------------------------------------

// The Jar eats itself to shield an artifact. The sacrifice is a cost,
// so it is gone before the ability resolves.
func TestWeldingJarSacrificesItselfToShieldAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Mana Rock", "Artifact", "test-regen-rock", false)
	jar := effectstest.PushCatalogPermanent(g, me.ID, "Welding Jar", "Artifact", effectstest.WeldingJarOracle, false)

	if err := g.ActivateCatalogAbility(me.ID, jar, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, jar); ok {
		t.Error("the sacrifice is a cost, paid at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.RegenShieldsOn(t, g, rock); n != 1 {
		t.Fatalf("shields on the artifact = %d, want 1", n)
	}
	effectstest.DestroyIt(t, g, rock)
	c, ok := effectstest.BattlefieldCard(g, rock)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	chirurgeon := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Chirurgeon", "Creature — Goblin Shaman", effectstest.GoblinChirurgeonOracle, false)
	fodder := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Token", TypeLine: "Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.SeedCreature(g, "Bear", me.ID)

	if err := g.ActivateCatalogAbility(me.ID, chirurgeon, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, fodder); ok {
		t.Error("the sacrificed Goblin is a cost, paid at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.RegenShieldsOn(t, g, bear); n != 1 {
		t.Fatalf("shields = %d, want 1", n)
	}
	effectstest.DestroyIt(t, g, bear)
	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Error("the shield should have replaced the destruction")
	}
}

// --- "can't be regenerated" ---------------------------------------

// The clause that was cosmetic on every card that printed it until
// there was a shield to ignore. Damnation kills through one;
// Day of Judgment, which does NOT print the clause, does not.
func TestDamnationKillsThroughAShieldAndDayOfJudgmentDoesNot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]

	// Day of Judgment first: the shield holds.
	survivor := effectstest.SeedCreature(g, "Bear", me.ID)
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(survivor); err != nil {
			t.Fatalf("RegenerateForEffect: %v", err)
		}
	})
	effectstest.CastCatalogSpell(t, g, "Day of Judgment", "Sorcery", effectstest.DayOfJudgmentOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, survivor); !ok {
		t.Fatal("Day of Judgment does not say they can't be regenerated, so the shield replaces its destruction")
	}

	// Damnation next, on a fresh shield: it does not.
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(survivor); err != nil {
			t.Fatalf("RegenerateForEffect: %v", err)
		}
	})
	effectstest.CastCatalogSpell(t, g, "Damnation", "Sorcery", effectstest.DamnationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, survivor); ok {
		t.Error("CR 701.19c: \"They can't be regenerated\" ignores the shield")
	}
}

// The single-target family carries the same rider — Terminate is the
// one that shares its whole body with Mortify and Putrefy.
func TestTerminateIgnoresARegenerationShield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	theirs := effectstest.SeedCreature(g, "Bear", them.ID)
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(theirs); err != nil {
			t.Fatalf("RegenerateForEffect: %v", err)
		}
	})
	_ = me

	effectstest.CastCatalogSpell(t, g, "Terminate", "Instant", effectstest.TerminateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)

	if _, ok := effectstest.BattlefieldCard(g, theirs); ok {
		t.Error("CR 701.19c: Terminate says it can't be regenerated")
	}
	if !them.Graveyard.Contains(theirs) {
		t.Error("the creature should be in its owner's graveyard")
	}
}
