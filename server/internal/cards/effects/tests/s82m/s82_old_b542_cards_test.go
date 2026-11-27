package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b542_cards_test.go — Shadowmoor, S82 OLD batch 542 (#5591).
// Every test plays the real catalog entry through the engine.

const (
	oldB542Feast  = "3ceae65e-938d-4e76-b3f6-7be70232549e"
	oldB542Mercy  = "4d80bac5-ce62-4c5c-873a-ce81d2697842"
	oldB542Sickle = "5b95a5ed-a6b2-4044-9273-2021d27794e9"
	oldB542Kite   = "9505f124-f152-48ce-b0d3-097abdda8e3e"
)

func TestOldB542EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB542Feast: "Gloomwidow's Feast", oldB542Mercy: "Mercy Killing",
		oldB542Sickle: "Blight Sickle", oldB542Kite: "Tatterkite",
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

func TestOldB542GloomwidowsFeastDestroysAFlyerAndMakesASpiderForBlueOrBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bat := effectstest.Pr7bCreature(g, opp.ID, "Dusk Bat", 1, "{1}{B}", []string{"B"}, "flying")
	effectstest.CastCatalogSpell(t, g, "Gloomwidow's Feast", "Instant", oldB542Feast, effectstest.CardRefs(bat))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bat) {
		t.Fatal("the black flyer survived")
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Spider"); got != 1 {
		t.Fatalf("the caster has %d Spiders, want 1 for a black creature", got)
	}
}

func TestOldB542GloomwidowsFeastMakesNoSpiderForAGreenFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := effectstest.Pr7bCreature(g, opp.ID, "Green Bird", 1, "{1}{G}", []string{"G"}, "flying")
	effectstest.CastCatalogSpell(t, g, "Gloomwidow's Feast", "Instant", oldB542Feast, effectstest.CardRefs(bird))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bird) {
		t.Fatal("the green flyer survived")
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Spider"); got != 0 {
		t.Fatalf("the caster has %d Spiders, want none for a green creature", got)
	}
}

func TestOldB542GloomwidowsFeastNeedsAFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Gloomwidow's Feast", "Instant", oldB542Feast, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("the Feast was cast at a creature without flying")
	}
}

func TestOldB542MercyKillingGivesTheControllerTokensEqualToPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Mercy Killing", "Instant", oldB542Mercy, effectstest.CardRefs(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Fatal("the Ogre was not sacrificed")
	}
	if got := effectstest.CountBattlefieldNamed(g, opp.ID, "Elf Warrior"); got != 3 {
		t.Errorf("the Ogre's controller has %d Elf Warriors, want 3", got)
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Elf Warrior"); got != 0 {
		t.Errorf("the caster has %d Elf Warriors, want none", got)
	}
}

func TestOldB542MercyKillingCountsPumpedPowerAndBeatsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	golem := effectstest.P1g2Indestructible(g, opp.ID, 2, 2)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(golem, game.CounterPlusOne, 2); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.CastCatalogSpell(t, g, "Mercy Killing", "Instant", oldB542Mercy, effectstest.CardRefs(golem))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(golem) {
		t.Fatal("the indestructible creature survived a sacrifice")
	}
	if got := effectstest.CountBattlefieldNamed(g, opp.ID, "Elf Warrior"); got != 4 {
		t.Errorf("the controller has %d Elf Warriors, want 4 (2 power plus two counters)", got)
	}
}

func TestOldB542BlightSickleGivesPlusOneAndWitherWhileEquipped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sickle := effectstest.PushCatalogPermanent(g, me.ID, "Blight Sickle", "Artifact — Equipment", oldB542Sickle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if effectstest.P1g2HasKeyword(g, bear, "wither") {
		t.Fatal("the Bear has wither before it is equipped")
	}
	if err := effectstest.StrictEquip(t, g, me.ID, sickle, bear, "{C}{C}"); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the equipped Bear is %d/%d, want 3/2", p, tough)
	}
	if !effectstest.P1g2HasKeyword(g, bear, "wither") {
		t.Error("the equipped Bear has no wither")
	}
}

func TestOldB542TatterkiteFliesAndTakesNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kite := effectstest.PushCatalogPermanent(g, me.ID, "Tatterkite", "Artifact Creature — Scarecrow", oldB542Kite, false)
	if !effectstest.P1g2HasKeyword(g, kite, "flying") {
		t.Error("Tatterkite has no flying")
	}
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(kite, game.CounterPlusOne, 2)
		_ = g.AddCounterForEffect(kite, game.CounterMinusOne, 1)
	})
	if got := effectstest.Pr8Counters(g, kite, game.CounterPlusOne) + effectstest.Pr8Counters(g, kite, game.CounterMinusOne); got != 0 {
		t.Errorf("Tatterkite has %d counters, want none", got)
	}
}
