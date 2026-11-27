package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b507_test.go — Conflux, batch 507 (#5587). Every test plays the real
// catalog entry through the engine.

const (
	oldB507Phoenix   = "67c50af9-bb8e-40b6-8b25-c209784b6865"
	oldB507Dragon    = "7f39bbae-6c84-4cb3-9a58-d746a723ad35"
	oldB507Meglonoth = "92e287b1-7cb1-4536-8942-d1ad059bf9bc"
	oldB507Fire      = "d2eeb1eb-38c0-4b76-ad96-a53b6bdc00e1"
	oldB507Cylix     = "47f20da0-cd31-4877-8eca-0c8389a66e89"
	oldB507Paragon   = "4e4991b3-11c4-4408-a752-67e51f0fae44"
	oldB507Brack     = "a8af4d61-1aed-415d-8d8e-7384194ee3e2"
	oldB507Avenger   = "e29350cd-c6ff-454d-a1fc-0c1bb60047b1"
	oldB507Bear      = "Creature — Bear"
)

func TestOldB507EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB507Phoenix: "Worldheart Phoenix", oldB507Dragon: "Voracious Dragon", oldB507Meglonoth: "Meglonoth",
		oldB507Fire: "Quenchable Fire", oldB507Cylix: "Mana Cylix", oldB507Paragon: "Paragon of the Amesha",
		oldB507Brack: "Brackwater Elemental", oldB507Avenger: "Giltspire Avenger",
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

func oldB507Phoenixes(g *game.Game, me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Worldheart Phoenix", TypeLine: "Creature — Phoenix",
		OracleID: oldB507Phoenix, ManaCost: "{3}{R}", Power: 2, Toughness: 2, Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	return id
}

func TestOldB507WorldheartPhoenixCastFromTheGraveyardEntersWithTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	phoenix := oldB507Phoenixes(g, me)
	effectstest.FloatMana(t, g, me, "{W}{U}{B}{R}{G}")
	if err := g.CastSpell(me.ID, phoenix, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: "worldheart-phoenix-graveyard"}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(phoenix) {
		t.Fatal("the Phoenix did not resolve onto the battlefield")
	}
	if got := effectstest.CounterCount(g, phoenix, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	if p, tt := effectstest.PtOf(t, g, phoenix); p != 4 || tt != 4 {
		t.Errorf("the Phoenix is %d/%d, want 4/4", p, tt)
	}
}

func TestOldB507WorldheartPhoenixCantBeCastFromTheGraveyardForItsManaCostAndHardCastHasNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	phoenix := oldB507Phoenixes(g, me)
	effectstest.FloatMana(t, g, me, "{R}{R}{R}{R}")
	if err := g.CastSpell(me.ID, phoenix, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Fatal("cast from the graveyard for its printed mana cost")
	}
	hand := effectstest.HandCardFull(me, "Worldheart Phoenix", "Creature — Phoenix", "{3}{R}", oldB507Phoenix, []string{"R"})
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{R}")
	if err := g.CastSpell(me.ID, hand, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("hard cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, hand, game.CounterPlusOne); got != 0 {
		t.Errorf("a hard-cast Phoenix has %d counters, want 0", got)
	}
}

func oldB507CastDragon(t *testing.T, g *game.Game, me *game.Player, devour ...uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.HandCardFull(me, "Voracious Dragon", "Creature — Dragon", "{3}{R}{R}", oldB507Dragon, []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, devour...)
	return id
}

func TestOldB507VoraciousDragonDealsTwiceTheGoblinsItDevoured(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gob1 := effectstest.B12Creature(g, me.ID, "Goblin One", "Creature — Goblin", 1, 1)
	gob2 := effectstest.B12Creature(g, me.ID, "Goblin Two", "Creature — Goblin Warrior", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB507Bear, 1, 1)
	before := effectstest.LifeOf(g, opp.ID)
	dragon := oldB507CastDragon(t, g, me, gob1, gob2, bear)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("pick target: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, dragon, game.CounterPlusOne); got != 3 {
		t.Errorf("the Dragon has %d counters, want 3 (devour 1, three creatures)", got)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != before-4 {
		t.Errorf("life %d, want %d (twice the two Goblins devoured)", got, before-4)
	}
}

func TestOldB507VoraciousDragonWithNoGoblinsDealsNoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB507Bear, 1, 1)
	before := effectstest.LifeOf(g, opp.ID)
	dragon := oldB507CastDragon(t, g, me, bear)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		_ = g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, dragon, game.CounterPlusOne); got != 1 {
		t.Errorf("the Dragon has %d counters, want 1", got)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != before {
		t.Errorf("life %d, want %d", got, before)
	}
}

func TestOldB507MeglonothDamagesTheBlockedCreaturesController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB507Bear, 2, 2)
	meg := effectstest.B12Push(g, opp.ID, "Meglonoth", "Creature — Beast", oldB507Meglonoth, 6, 6)
	before := effectstest.LifeOf(g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(meg, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != before-6 {
		t.Errorf("the attacker's controller has %d life, want %d (6 from Meglonoth's power)", got, before-6)
	}
}

func TestOldB507ManaCylixFiltersOneManaIntoAnyColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cylix := effectstest.PushCatalogPermanent(g, me.ID, "Mana Cylix", "Artifact", oldB507Cylix, false)
	effectstest.AdvanceToMain(t, g)
	spec, _ := effects.Lookup(oldB507Cylix)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Cost.Mana != "{1}" || !spec.ManaAbilities[0].Cost.Tap {
		t.Fatalf("mana abilities %+v, want one {1},{T} ability", spec.ManaAbilities)
	}
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.ActivateManaFor(t, g, me.ID, cylix, 0, game.ManaAbilityParams{})
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil {
		t.Fatal("the Cylix did not ask which colour to add")
	}
	for _, want := range []string{"W", "U", "B", "R", "G"} {
		found := false
		for _, o := range pick.ColorOptions {
			found = found || o == want
		}
		if !found {
			t.Errorf("the colour %s is not offered: %v", want, pick.ColorOptions)
		}
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("pool holds %d mana, want 1 (the {C} was spent, one mana added)", got)
	} else if me.ManaPool.CountColor("G") != 1 {
		t.Error("the mana added is not green")
	}
	if !effectstest.TappedNow(g, cylix) {
		t.Error("the Cylix is untapped")
	}
}

func TestOldB507ParagonOfTheAmeshaBecomesAFlyingLifelinkingAngel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	paragon := effectstest.B12Push(g, me.ID, "Paragon of the Amesha", "Creature — Human Knight", oldB507Paragon, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "WUBRG")
	if err := g.ActivateCatalogAbility(me.ID, paragon, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, paragon); p != 5 || tt != 5 {
		t.Errorf("the Paragon is %d/%d, want 5/5", p, tt)
	}
	for _, kw := range []string{"flying", "lifelink", "first strike"} {
		if !effectstest.HasEffectiveKeyword(t, g, paragon, kw) {
			t.Errorf("the Paragon lacks %s", kw)
		}
	}
	c, _ := g.LookupCardForEffect(paragon)
	if !c.HasSubtypeWord("Angel") || c.HasSubtypeWord("Human") {
		t.Errorf("subtypes %v, want only Angel", c.Effective().Subtypes)
	}
}

func TestOldB507BrackwaterElementalIsSacrificedAtTheNextEndStepAfterAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elem := effectstest.B12Push(g, me.ID, "Brackwater Elemental", "Creature — Elemental", oldB507Brack, 4, 4)
	effectstest.DeclareAttack(t, g, opp.ID, elem)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(elem) {
		t.Fatal("the Elemental was sacrificed before the end step")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(elem) {
		t.Error("the Elemental is still on the battlefield after the end step")
	}
	if z := effectstest.ZoneOfCard(g, elem); z != game.ZoneGraveyard {
		t.Errorf("the Elemental is in zone %v, want the graveyard", z)
	}
}

func TestOldB507BrackwaterElementalHasUnearth(t *testing.T) {
	spec, _ := effects.Lookup(oldB507Brack)
	found := false
	for _, a := range spec.Activated {
		if len(a.Zones) == 1 && a.Zones[0] == game.ZoneGraveyard && a.SorcerySpeed {
			found = true
		}
	}
	if !found {
		t.Error("no sorcery-speed graveyard ability (unearth)")
	}
}

func TestOldB507GiltspireAvengerDestroysACreatureThatDamagedYouThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	avenger := effectstest.B12Push(g, me.ID, "Giltspire Avenger", "Creature — Human Soldier", oldB507Avenger, 2, 2)
	hitter := effectstest.B12Creature(g, opp.ID, "Hitter", oldB507Bear, 2, 2)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", oldB507Bear, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, avenger, "exalted") {
		t.Error("the Avenger has no exalted")
	}
	// No creature has damaged you yet: nothing is a legal target.
	if err := g.ActivateCatalogAbility(me.ID, avenger, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.CardTargets(hitter)}); err == nil {
		t.Fatal("targeted a creature that had not damaged you")
	}
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(hitter, me.ID, 2)
	})
	if err := g.ActivateCatalogAbility(me.ID, avenger, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.CardTargets(bystander)}); err == nil {
		t.Fatal("targeted a creature that did not damage you")
	}
	if err := g.ActivateCatalogAbility(me.ID, avenger, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.CardTargets(hitter)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hitter) {
		t.Error("the creature that dealt damage to you was not destroyed")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("the bystander was destroyed")
	}
}

func oldB507QuenchableFire(t *testing.T, g *game.Game, target game.TargetRef) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Quenchable Fire", "Sorcery", oldB507Fire, []game.TargetRef{target})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB507QuenchableFireDealsMoreDamageNextUpkeepUnlessPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	before := effectstest.LifeOf(g, opp.ID)
	oldB507QuenchableFire(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if got := effectstest.LifeOf(g, opp.ID); got != before-3 {
		t.Fatalf("life %d after the spell, want %d", got, before-3)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-6 {
		t.Errorf("life %d after the upkeep, want %d (3 more, unpaid)", got, before-6)
	}
}

func TestOldB507QuenchableFirePaidWithBlueDealsNoMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.SeedLand(g, opp.ID, "Island", "Basic Land — Island", "")
	before := effectstest.LifeOf(g, opp.ID)
	oldB507QuenchableFire(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	var id uuid.UUID
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.PayToCancel != nil {
			id = dt.ID
		}
	}
	if id == uuid.Nil {
		t.Fatal("no delayed trigger with a payment")
	}
	if err := g.PerformSpecialAction(opp.ID, id, game.SpecialActionPayToCancel, game.SpecialActionParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("pay {U}: %v", err)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-3 {
		t.Errorf("life %d, want %d (the {U} was paid)", got, before-3)
	}
}

func TestOldB507QuenchableFireHitsAPlaneswalkerItsControllerMayPayFor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Walker",
		TypeLine: "Legendary Planeswalker — Test", Counters: map[string]int{game.CounterLoyalty: 9}, Owner: opp.ID, Controller: opp.ID})
	oldB507QuenchableFire(t, g, game.TargetRef{Kind: game.TargetCard, ID: walker})
	var offered bool
	g.WithWriteLock(func() { offered = len(g.PayToCancelOffersLocked(opp.ID)) == 1 })
	if !offered {
		t.Error("the planeswalker's controller is not offered the {U} payment")
	}
	if got := effectstest.CounterCount(g, walker, game.CounterLoyalty); got != 6 {
		t.Errorf("loyalty %d, want 6 after 3 damage", got)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, walker, game.CounterLoyalty); got != 3 {
		t.Errorf("loyalty %d, want 3 after 3 more damage", got)
	}
}
