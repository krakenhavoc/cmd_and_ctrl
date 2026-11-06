package effects

import (
	"testing"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Rabanastre, Royal City ---------------------------------------------------------

func TestFinRURabanastreEntersTappedAndTapsForRedOrWhite(t *testing.T) {
	g := newCatalogGame(t)
	land := playLandFromHand(t, g, "Rabanastre, Royal City", finRURabanastreOracle)
	if c := findBattlefieldCardByID(g, land); c == nil || !c.Tapped {
		t.Error("Rabanastre should enter tapped")
	}
	spec, _ := Lookup(finRURabanastreOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{R|W}" || !spec.ManaAbilities[0].Cost.Tap {
		t.Errorf("mana abilities %+v, want {T}: {R|W}", spec.ManaAbilities)
	}
}

// --- Red Mage's Rapier --------------------------------------------------------------

func TestFinRURedMagesRapierHeroGrowsOnNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rapier := castCatalogSpell(t, g, "Red Mage's Rapier", "Artifact — Equipment", finRURapierOracle, nil)
	passPriorityAroundTable(t, g)
	hero := ysBHero(t, g, rapier)
	finMPHeroPT(t, g, hero, 1, 1, "Wizard")
	finMPEquipLabel(t, finRURapierOracle, "{3}")

	advanceToMain(t, g)
	bear := socB17Hand(me, "Test Bear", "Creature — Bear", "{1}{G}", socB17PlainSorceryOracle)
	floatForTest(g, me, "GG")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	finMPHeroPT(t, g, hero, 1, 1, "Wizard")

	sorcery := socB17Hand(me, "Test Spell", "Sorcery", "{R}", socB17PlainSorceryOracle)
	floatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the sorcery: %v", err)
	}
	passPriorityAroundTable(t, g)
	finMPHeroPT(t, g, hero, 3, 1, "Wizard")
}

// --- Rook Turret --------------------------------------------------------------------

func TestFinRURookTurretLootsWhenAnotherArtifactOfYoursEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc18Library(me, 6)
	turret := b12Push(g, me.ID, "Rook Turret", "Artifact Creature — Construct", finRUTurretOracle, 3, 3)
	if !hasEffectiveKeyword(t, g, turret, "flying") {
		t.Error("the Turret has no flying")
	}

	avB1OpponentCreatureEnters(t, g, opp.ID, "Artifact — Clue")
	if ask := latestChoiceOfKind(g, game.PendingChoiceConfirm); ask != nil {
		t.Fatalf("an opponent's artifact asked to loot: %+v", ask)
	}
	avB1OpponentCreatureEnters(t, g, me.ID, "Creature — Bear")
	if ask := latestChoiceOfKind(g, game.PendingChoiceConfirm); ask != nil {
		t.Fatalf("a nonartifact creature asked to loot: %+v", ask)
	}

	avB1OpponentCreatureEnters(t, g, me.ID, "Artifact — Clue")
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
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
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after the discard, want %d", me.Hand.Size(), hand)
	}
	passPriorityAroundTable(t, g)

	// Declining draws nothing and discards nothing.
	avB1OpponentCreatureEnters(t, g, me.ID, "Artifact — Clue")
	ask = latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("no prompt for the second artifact")
	}
	hand = me.Hand.Size()
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm(decline): %v", err)
	}
	if me.Hand.Size() != hand || discardChoiceFor(g, me.ID) != nil {
		t.Errorf("declining changed the hand (%d → %d) or asked for a discard", hand, me.Hand.Size())
	}
}

// --- Rosa, Resolute White Mage ------------------------------------------------------

func TestFinRURosaCountersAndGivesLifelinkAtYourCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rosa := pushCatalogPermanent(g, me.ID, "Rosa, Resolute White Mage", "Legendary Creature — Human Noble Cleric", finRURosaOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, rosa, "reach") {
		t.Error("Rosa has no reach")
	}

	advanceTo(t, g, game.StepBeginCombat)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	if !hasID(prompt.PickTargetCards, rosa) || !hasID(prompt.PickTargetCards, bear) || hasID(prompt.PickTargetCards, theirs) {
		t.Errorf("legal targets %v: want Rosa and the Bear, not the opponent's creature", prompt.PickTargetCards)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("the Bear did not gain lifelink")
	}
	if hasEffectiveKeyword(t, g, rosa, "lifelink") {
		t.Error("Rosa gained lifelink without being the target")
	}
}

// --- Ultimecia, Temporal Threat -----------------------------------------------------

func TestFinRUUltimeciaTapsTheirCreaturesAndDrawsOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc18Library(me, 6)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, g.Seats[2].ID, "Another Bear", "Creature — Bear", 2, 2)
	castRingCreature(t, g, "Ultimecia, Temporal Threat", "Legendary Creature — Human Warlock", finRUUltimeciaOracle, 4, 4)
	passPriorityAroundTable(t, g)
	if !isTapped(g, theirs) || !isTapped(g, other) {
		t.Error("every opponent's creature should be tapped")
	}
	if isTapped(g, mine) {
		t.Error("your own creature was tapped")
	}

	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, mine, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn for your creature's combat damage", hand, me.Hand.Size())
	}

	hand = me.Hand.Size()
	dealCombatDamageToPlayer(g, theirs, me.ID, 2)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("an opponent's creature's damage drew you a card: hand %d → %d", hand, me.Hand.Size())
	}
}
