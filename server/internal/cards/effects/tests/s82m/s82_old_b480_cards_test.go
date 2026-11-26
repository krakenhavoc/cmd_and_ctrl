package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b480_cards_test.go — Rise of the Eldrazi: activated abilities,
// batch 480 (#5581, S82 OLD). Every test plays the real catalog entry through
// the engine.

const (
	oldB480Slinger = "1a852283-42d4-4197-91b9-5c952c7325ca"
	oldB480Invoker = "847587eb-8713-496e-b0c7-41ad6d43b852"
	oldB480Bulwark = "8916b3d4-6a9a-420a-8ee7-be67b2f47201"
	oldB480Nimbus  = "ab5b49d4-e648-43f1-bcc0-7dd4763316ee"
	oldB480Giganto = "c4b5567a-f195-4638-94b5-0b933dba1850"
	oldB480Wurm    = "d988ed15-2927-421b-9a6e-01f0261272c2"
)

func TestOldB480CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB480Slinger: "Kor Line-Slinger",
		oldB480Invoker: "Frostwind Invoker",
		oldB480Bulwark: "Reinforced Bulwark",
		oldB480Nimbus:  "Rage Nimbus",
		oldB480Giganto: "Gigantomancer",
		oldB480Wurm:    "Skeletal Wurm",
	} {
		effectstest.RequireFullNoCaveats(t, oracle, name)
	}
}

func TestOldB480KorLineSlingerTapsACreatureWithPowerThreeOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	slinger := effectstest.B12Push(g, me.ID, "Kor Line-Slinger", "Creature — Kor Scout", oldB480Slinger, 0, 1)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 4, 4)
	if err := g.ActivateCatalogAbility(me.ID, slinger, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(big)}); err == nil {
		t.Fatal("Line-Slinger targeted a creature with power 4")
	}
	effectstest.B16Activate(t, g, me.ID, slinger, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(small)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, small) {
		t.Error("the target is untapped")
	}
	if !effectstest.IsTapped(g, slinger) {
		t.Error("the Line-Slinger did not tap for its own cost")
	}
}

func TestOldB480FrostwindInvokerGivesYourCreaturesFlyingForEightMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	invoker := effectstest.B12Push(g, me.ID, "Frostwind Invoker", "Creature — Merfolk Wizard", oldB480Invoker, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, invoker, "flying") {
		t.Error("no flying")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, invoker, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for seven mana")
	}
	effectstest.AddPool(t, g, me.ID, "{C}")
	effectstest.B16Activate(t, g, me.ID, invoker, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("my Bear did not gain flying")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "flying") {
		t.Error("an opponent's creature gained flying")
	}
	late := effectstest.B12Creature(g, me.ID, "Latecomer", "Creature — Bear", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, late, "flying") {
		t.Error("a creature that arrived after resolution gained flying")
	}
}

func TestOldB480ReinforcedBulwarkPreventsTheNextDamageToYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	wall := effectstest.B12Push(g, opp.ID, "Reinforced Bulwark", "Artifact Creature — Wall", oldB480Bulwark, 0, 4)
	if !effectstest.HasEffectiveKeyword(t, g, wall, "defender") {
		t.Error("no defender")
	}
	// The shield is the Wall's controller's, here the opponent (seat 1), who
	// activates it with seat 0 the active player.
	effectstest.B16Activate(t, g, opp.ID, wall, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 2 {
		t.Errorf("the controller lost %d life to a Lightning Bolt, want 2 (1 prevented)", got)
	}
}

func TestOldB480RageNimbusForcesACreatureToAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	nimbus := effectstest.B12Push(g, me.ID, "Rage Nimbus", "Creature — Elemental", oldB480Nimbus, 5, 3)
	for _, kw := range []string{"defender", "flying"} {
		if !effectstest.HasEffectiveKeyword(t, g, nimbus, kw) {
			t.Errorf("no %s", kw)
		}
	}
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, Keywords: []string{"haste"}})
	effectstest.AddPool(t, g, me.ID, "{C}{R}")
	effectstest.B16Activate(t, g, me.ID, nimbus, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	g.WithWriteLock(func() {
		if g.AttackRequirementsUnmetForEffect() == nil {
			t.Error("the targeted creature owes an attack but nothing says so")
		}
	})
}

func TestOldB480GigantomancerSetsBasePowerAndToughnessToSevenUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	giant := effectstest.B12Push(g, me.ID, "Gigantomancer", "Creature — Human Shaman", oldB480Giganto, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AddPool(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, giant, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(theirs)}); err == nil {
		t.Fatal("Gigantomancer targeted an opponent's creature")
	}
	effectstest.B16Activate(t, g, me.ID, giant, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 7 || tough != 7 {
		t.Errorf("the Bear is %d/%d, want 7/7", p, tough)
	}
	// A +1/+1 counter still lands on top of the base (CR 613.4).
	effectstest.PutCounters(t, g, bear, game.CounterPlusOne, 1)
	if p, tough := effectstest.PtOf(t, g, bear); p != 8 || tough != 8 {
		t.Errorf("the Bear with a counter is %d/%d, want 8/8", p, tough)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("after the turn the Bear is %d/%d, want 3/3", p, tough)
	}
}

func TestOldB480SkeletalWurmRegeneratesForOneBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	wurm := effectstest.B12Push(g, me.ID, "Skeletal Wurm", "Creature — Skeleton Wurm", oldB480Wurm, 7, 6)
	effectstest.AddPool(t, g, me.ID, "{B}")
	effectstest.B16Activate(t, g, me.ID, wurm, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.RegenShieldsOn(t, g, wurm); n != 1 {
		t.Fatalf("%d regeneration shields, want 1", n)
	}
	effectstest.DestroyIt(t, g, wurm)
	if !g.Battlefield.Contains(wurm) {
		t.Error("the regenerated Wurm was destroyed")
	}
}
