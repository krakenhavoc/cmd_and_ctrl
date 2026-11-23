package testsex

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_ru_cards_test.go — Final Fantasy (FIN), cards R and U (#4462).
// Each test plays the real catalog entry and asserts what its text says.

const (
	finRURabanastreOracle = "bd39e333-b0b8-4a85-8645-d77f4a3941db"
	finRURapierOracle     = "35ad9e15-3e00-4bd2-96cf-36a6ef49059f"
	finRUTurretOracle     = "c5a6d752-9975-4059-8552-55c6ffb3ab68"
	finRURosaOracle       = "9c4fdd20-0396-40b7-8687-43b1035b6956"
	finRUUltimeciaOracle  = "5a8e7e35-e9ea-4808-8aa7-7580a23b9d9d"
)

func TestFinRUEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		finRURabanastreOracle: "Rabanastre, Royal City",
		finRURapierOracle:     "Red Mage's Rapier",
		finRUTurretOracle:     "Rook Turret",
		finRURosaOracle:       "Rosa, Resolute White Mage",
		finRUUltimeciaOracle:  "Ultimecia, Temporal Threat",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Rabanastre, Royal City ---------------------------------------------------------

func TestFinRURabanastreEntersTappedAndTapsForRedOrWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.PlayLandFromHand(t, g, "Rabanastre, Royal City", finRURabanastreOracle)
	if c := effectstest.FindBattlefieldCardByID(g, land); c == nil || !c.Tapped {
		t.Error("Rabanastre should enter tapped")
	}
	spec, _ := effects.Lookup(finRURabanastreOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{R|W}" || !spec.ManaAbilities[0].Cost.Tap {
		t.Errorf("mana abilities %+v, want {T}: {R|W}", spec.ManaAbilities)
	}
}

// --- Red Mage's Rapier --------------------------------------------------------------

func TestFinRURedMagesRapierHeroGrowsOnNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rapier := effectstest.CastCatalogSpell(t, g, "Red Mage's Rapier", "Artifact — Equipment", finRURapierOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hero := effectstest.YsBHero(t, g, rapier)
	finMPHeroPT(t, g, hero, 1, 1, "Wizard")
	finMPEquipLabel(t, finRURapierOracle, "{3}")

	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SocB17Hand(me, "Test Bear", "Creature — Bear", "{1}{G}", effectstest.SocB17PlainSorceryOracle)
	effectstest.FloatForTest(g, me, "GG")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	finMPHeroPT(t, g, hero, 1, 1, "Wizard")

	sorcery := effectstest.SocB17Hand(me, "Test Spell", "Sorcery", "{R}", effectstest.SocB17PlainSorceryOracle)
	effectstest.FloatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the sorcery: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	finMPHeroPT(t, g, hero, 3, 1, "Wizard")
}

// --- Rook Turret --------------------------------------------------------------------

func TestFinRURookTurretLootsWhenAnotherArtifactOfYoursEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc18Library(me, 6)
	turret := effectstest.B12Push(g, me.ID, "Rook Turret", "Artifact Creature — Construct", finRUTurretOracle, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, turret, "flying") {
		t.Error("the Turret has no flying")
	}

	effectstest.AvB1OpponentCreatureEnters(t, g, opp.ID, "Artifact — Clue")
	if ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm); ask != nil {
		t.Fatalf("an opponent's artifact asked to loot: %+v", ask)
	}
	effectstest.AvB1OpponentCreatureEnters(t, g, me.ID, "Creature — Bear")
	if ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm); ask != nil {
		t.Fatalf("a nonartifact creature asked to loot: %+v", ask)
	}

	effectstest.AvB1OpponentCreatureEnters(t, g, me.ID, "Artifact — Clue")
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatal("no \"draw a card?\" prompt when your artifact entered")
	}
	hand := me.Hand.Size()
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d after drawing, want %d", me.Hand.Size(), hand+1)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after the discard, want %d", me.Hand.Size(), hand)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// Declining draws nothing and discards nothing.
	effectstest.AvB1OpponentCreatureEnters(t, g, me.ID, "Artifact — Clue")
	ask = effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("no prompt for the second artifact")
	}
	hand = me.Hand.Size()
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm(decline): %v", err)
	}
	if me.Hand.Size() != hand || effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Errorf("declining changed the hand (%d → %d) or asked for a discard", hand, me.Hand.Size())
	}
}

// --- Rosa, Resolute White Mage ------------------------------------------------------

func TestFinRURosaCountersAndGivesLifelinkAtYourCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rosa := effectstest.PushCatalogPermanent(g, me.ID, "Rosa, Resolute White Mage", "Legendary Creature — Human Noble Cleric", finRURosaOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, rosa, "reach") {
		t.Error("Rosa has no reach")
	}

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	if !effectstest.HasID(prompt.PickTargetCards, rosa) || !effectstest.HasID(prompt.PickTargetCards, bear) || effectstest.HasID(prompt.PickTargetCards, theirs) {
		t.Errorf("legal targets %v: want Rosa and the Bear, not the opponent's creature", prompt.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("the Bear did not gain lifelink")
	}
	if effectstest.HasEffectiveKeyword(t, g, rosa, "lifelink") {
		t.Error("Rosa gained lifelink without being the target")
	}
}

// --- Ultimecia, Temporal Threat -----------------------------------------------------

func TestFinRUUltimeciaTapsTheirCreaturesAndDrawsOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc18Library(me, 6)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, g.Seats[2].ID, "Another Bear", "Creature — Bear", 2, 2)
	effectstest.CastRingCreature(t, g, "Ultimecia, Temporal Threat", "Legendary Creature — Human Warlock", finRUUltimeciaOracle, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, theirs) || !effectstest.IsTapped(g, other) {
		t.Error("every opponent's creature should be tapped")
	}
	if effectstest.IsTapped(g, mine) {
		t.Error("your own creature was tapped")
	}

	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, mine, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn for your creature's combat damage", hand, me.Hand.Size())
	}

	hand = me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, theirs, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("an opponent's creature's damage drew you a card: hand %d → %d", hand, me.Hand.Size())
	}
}
