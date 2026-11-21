package testsm

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Multiverse Gift Box (#5632): Bull Elephant, Dark Privilege, King
// Cheetah, Necrosavant, Ovinomancer, Urborg Mindsucker, Vampirism,
// Viashino Sandstalker and Wicked Reward.

const (
	bullElephantOracle        = "4ed04ece-64d2-41a4-bf36-b993cbbf2aac"
	darkPrivilegeOracle       = "29f4c69a-ca59-4e9f-a808-87ba85b34f2f"
	kingCheetahOracle         = "f9ca35e1-24c5-46f4-a666-0dfe2c920e8b"
	necrosavantOracle         = "6732a587-e3cc-4c30-9267-f8949dd8724a"
	ovinomancerOracle         = "98026d30-a29f-4590-968f-b1b36926a275"
	urborgMindsuckerOracle    = "19d11ca8-a95b-4e6e-a71f-8f9d06c580fd"
	vampirismOracle           = "14df239a-c7cc-492a-94f5-2971f6efe17b"
	viashinoSandstalkerOracle = "cbfccfc0-3117-4ca0-b7ce-8e2d67ec49ab"
	wickedRewardOracle        = "7aad0a80-cf7b-4224-8e72-d611e0be3383"
)

// mgbCastCreature casts a catalog creature from the active player's hand
// in a main phase and lets the spell and its enters trigger resolve.
func mgbCastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	g.WithWriteLock(func() {
		active.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
		})
	})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// --- Bull Elephant ----------------------------------------------------

// One Forest can't pay for two: the Elephant is sacrificed and the Forest
// stays.
func TestBullElephantIsSacrificedWithOnlyOneForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.MgbLand(g, me.ID, "Forest", "Basic Land — Forest")
	elephant := mgbCastCreature(t, g, "Bull Elephant", "Creature — Elephant", bullElephantOracle, 4, 4)
	if g.Battlefield.Contains(elephant) {
		t.Fatal("Bull Elephant stayed with one Forest to return")
	}
	if !g.Battlefield.Contains(forest) {
		t.Error("the lone Forest was returned anyway")
	}
}

// Two Forests are offered (a nonbasic Forest included, a Mountain not);
// one is not a payment; both keep the Elephant.
func TestBullElephantReturnsTwoForestsOrIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.MgbLand(g, me.ID, "Forest", "Basic Land — Forest")
	trop := effectstest.MgbLand(g, me.ID, "Tropical Island", "Land — Forest Island")
	mountain := effectstest.MgbLand(g, me.ID, "Mountain", "Basic Land — Mountain")
	elephant := mgbCastCreature(t, g, "Bull Elephant", "Creature — Elephant", bullElephantOracle, 4, 4)

	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to return two Forests")
	}
	if !slices.Contains(pick.ChooseCards, forest) || !slices.Contains(pick.ChooseCards, trop) || slices.Contains(pick.ChooseCards, mountain) {
		t.Fatalf("offered %v; want the two Forests and not the Mountain", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{forest}); err == nil {
		t.Fatal("one Forest was accepted as the payment")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{forest, trop}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(elephant) {
		t.Error("Bull Elephant was sacrificed after two Forests were returned")
	}
	if !me.Hand.Contains(forest) || !me.Hand.Contains(trop) {
		t.Error("the two Forests are not in their owner's hand")
	}
	if !g.Battlefield.Contains(mountain) {
		t.Error("the Mountain moved")
	}

	// Declining sacrifices it and keeps the Forests.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	f1 := effectstest.MgbLand(g2, me2.ID, "Forest", "Basic Land — Forest")
	f2 := effectstest.MgbLand(g2, me2.ID, "Forest", "Basic Land — Forest")
	elephant2 := mgbCastCreature(t, g2, "Bull Elephant", "Creature — Elephant", bullElephantOracle, 4, 4)
	pick2 := effectstest.LatestChooseCardsFor(g2, me2.ID)
	if pick2 == nil {
		t.Fatal("no prompt")
	}
	if err := g2.ResolveChooseCards(pick2.ID, me2.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards(none): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(elephant2) {
		t.Error("declining kept Bull Elephant")
	}
	if !g2.Battlefield.Contains(f1) || !g2.Battlefield.Contains(f2) {
		t.Error("declining returned a Forest")
	}
}

// --- Ovinomancer ------------------------------------------------------

// Three basic lands are the payment; a nonbasic land is not offered.
func TestOvinomancerReturnsThreeBasicLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	island := effectstest.MgbLand(g, me.ID, "Island", "Basic Land — Island")
	swamp := effectstest.MgbLand(g, me.ID, "Swamp", "Basic Land — Swamp")
	plains := effectstest.MgbLand(g, me.ID, "Plains", "Basic Land — Plains")
	trop := effectstest.MgbLand(g, me.ID, "Tropical Island", "Land — Forest Island")
	ovino := mgbCastCreature(t, g, "Ovinomancer", "Creature — Human Wizard", ovinomancerOracle, 0, 1)

	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to return three basic lands")
	}
	if slices.Contains(pick.ChooseCards, trop) || len(pick.ChooseCards) != 3 {
		t.Fatalf("offered %v; want the three basic lands only", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{island, swamp, plains}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ovino) {
		t.Error("Ovinomancer was sacrificed after three basic lands were returned")
	}
	for _, id := range []uuid.UUID{island, swamp, plains} {
		if !me.Hand.Contains(id) {
			t.Error("a basic land is not in its owner's hand")
		}
	}
	if !g.Battlefield.Contains(trop) {
		t.Error("the nonbasic land moved")
	}

	// Two basic lands are too few: sacrificed at once.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.MgbLand(g2, me2.ID, "Island", "Basic Land — Island")
	effectstest.MgbLand(g2, me2.ID, "Swamp", "Basic Land — Swamp")
	ovino2 := mgbCastCreature(t, g2, "Ovinomancer", "Creature — Human Wizard", ovinomancerOracle, 0, 1)
	if g2.Battlefield.Contains(ovino2) {
		t.Error("Ovinomancer stayed with two basic lands")
	}
}

// {T}, return it to hand: the target is destroyed through a regeneration
// shield, and its controller gets a 0/1 green Sheep.
func TestOvinomancerDestroysACreatureAndGivesItsControllerASheep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ovino := effectstest.PushCatalogPermanent(g, me.ID, "Ovinomancer", "Creature — Human Wizard", ovinomancerOracle, false)
	target := mm2Creature(g, opp.ID, "Their Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(target); err != nil {
			t.Fatalf("shield: %v", err)
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, ovino, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(ovino) || !me.Hand.Contains(ovino) {
		t.Error("Ovinomancer was not returned to hand as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(target) {
		t.Error("the target was regenerated; it can't be")
	}
	var sheep *game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Sheep" {
				sheep = &c
			}
		}
	})
	if sheep == nil {
		t.Fatal("no Sheep token")
	}
	if sheep.Controller != opp.ID || sheep.Power != 0 || sheep.Toughness != 1 || !slices.Contains(sheep.Colors, "G") {
		t.Errorf("Sheep: controller %v (want the target's), %d/%d %v; want a 0/1 green", sheep.Controller, sheep.Power, sheep.Toughness, sheep.Colors)
	}
}

// A creature that arrived this turn can't pay {T} (CR 302.6).
func TestOvinomancerCantTapTheTurnItArrives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ovino := effectstest.PushCatalogPermanent(g, me.ID, "Ovinomancer", "Creature — Human Wizard", ovinomancerOracle, true)
	target := mm2Creature(g, opp.ID, "Their Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, ovino, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err == nil {
		t.Fatal("a summoning-sick Ovinomancer paid {T}")
	}
	if !g.Battlefield.Contains(ovino) {
		t.Error("a refused activation returned Ovinomancer")
	}
}

// --- Dark Privilege ---------------------------------------------------

// +1/+1, and sacrificing another creature shields the enchanted one.
func TestDarkPrivilegePumpsAndRegeneratesTheEnchantedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := mm2Creature(g, me.ID, "Bear", 2, 2)
	fodder := mm2Creature(g, me.ID, "Fodder", 1, 1)
	aura := effectstest.AuraCast(t, g, "Dark Privilege", darkPrivilegeOracle, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Fatalf("enchanted Bear is %d/%d, want 3/3", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice was not paid at activation")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.RegenShieldsOn(t, g, bear); n != 1 {
		t.Fatalf("shields = %d, want 1", n)
	}
	effectstest.DestroyIt(t, g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Error("the regenerated Bear was destroyed")
	}
}

// --- King Cheetah -----------------------------------------------------

// Flash: cast in the draw step, which a creature without it can't be.
func TestKingCheetahHasFlash(t *testing.T) {
	g2 := effectstest.NewCatalogGame(t) // parked on seat 0's draw step
	me := g2.Seats[0]
	cheetah := effectstest.HandCardOf(g2, me, "King Cheetah", "Creature — Cat", "{3}{G}", kingCheetahOracle)
	bear := effectstest.HandCardOf(g2, me, "Bear", "Creature — Bear", "{3}{G}", "")
	g2.WithWriteLock(func() { effectstest.FillPoolColored(me, "G", 4) })
	if err := g2.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a creature without flash was cast in the draw step")
	}
	if err := g2.CastSpell(me.ID, cheetah, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("King Cheetah in the draw step: %v", err)
	}
}

// --- Necrosavant ------------------------------------------------------

// From the graveyard, sacrificing a creature, only in your own upkeep.
func TestNecrosavantReturnsDuringYourUpkeepOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	necro := effectstest.PushCatalogGraveyardCard(me, "Necrosavant", "Creature — Zombie Giant", necrosavantOracle, 5, 5)
	fodder := mm2Creature(g, me.ID, "Fodder", 1, 1)
	params := game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}
	fund := func() {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{B}{B}{B}{B}"); err != nil {
			t.Fatalf("fund: %v", err)
		}
	}

	fund() // seat 0's draw step: its own turn, wrong step
	effectstest.AcRefused(t, g, me.ID, necro, 0, params)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	fund()
	effectstest.AcRefused(t, g, me.ID, necro, 0, params)

	effectstest.AdvanceToUpkeepOf(t, g, 0)
	fund()
	if err := g.ActivateCatalogAbility(me.ID, necro, 0, params); err != nil {
		t.Fatalf("activate in your upkeep: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice was not paid at activation")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(necro) || me.Graveyard.Contains(necro) {
		t.Error("Necrosavant did not return to the battlefield")
	}
}

// --- Urborg Mindsucker ------------------------------------------------

// Sacrificed as the cost, at sorcery speed only; the opponent discards one
// card at random.
func TestUrborgMindsuckerMakesAnOpponentDiscardAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sucker := effectstest.PushCatalogPermanent(g, me.ID, "Urborg Mindsucker", "Creature — Horror", urborgMindsuckerOracle, false)
	params := game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}

	// The draw step is not a sorcery's window.
	if err := g.ActivateCatalogAbility(me.ID, sucker, 0, params); err == nil {
		t.Fatal("activated outside a main phase")
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sucker, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err == nil {
		t.Fatal("targeted its own controller, not an opponent")
	}
	before := opp.Hand.Size()
	if before == 0 {
		t.Fatal("the opponent has no cards to discard")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}"); err != nil {
		t.Fatalf("fund: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, sucker, 0, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(sucker) {
		t.Error("the Mindsucker was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != before-1 {
		t.Errorf("opponent's hand %d -> %d, want one card discarded", before, got)
	}
}

// --- Vampirism --------------------------------------------------------

// +1/+1 for each other creature you control, -1/-1 to those others, and a
// card drawn at the next turn's upkeep.
func TestVampirismPumpsTheHostShrinksTheOthersAndDrawsLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := mm2Creature(g, me.ID, "Bear", 2, 2)
	a := mm2Creature(g, me.ID, "Ally A", 2, 2)
	b := mm2Creature(g, me.ID, "Ally B", 2, 2)
	effectstest.AuraCast(t, g, "Vampirism", vampirismOracle, bear)

	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("enchanted Bear is %d/%d, want 4/4 (two other creatures)", p, tough)
	}
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != 1 || tough != 1 {
			t.Errorf("another creature is %d/%d, want 1/1", p, tough)
		}
	}

	hand := me.Hand.Size()
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d at the next turn's upkeep, want one card drawn", hand, got)
	}
}

// On an opponent's creature: it counts the Aura controller's creatures, and
// the Aura controller's creatures shrink, the opponent's others do not.
func TestVampirismOnAnOpponentsCreatureCountsAndShrinksYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := mm2Creature(g, opp.ID, "Their Bear", 2, 2)
	theirOther := mm2Creature(g, opp.ID, "Their Other", 2, 2)
	mine := mm2Creature(g, me.ID, "My Bear", 3, 3)
	effectstest.AuraCast(t, g, "Vampirism", vampirismOracle, theirs)

	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 3 || tough != 3 {
		t.Errorf("enchanted creature is %d/%d, want 3/3 (one creature you control)", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("your creature is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirOther); p != 2 || tough != 2 {
		t.Errorf("the opponent's other creature is %d/%d, want 2/2", p, tough)
	}
}

// --- Viashino Sandstalker ---------------------------------------------

// Haste, and back to hand at the end step: yours or another player's.
func TestViashinoSandstalkerReturnsAtEachEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	lizard := effectstest.PushCatalogPermanent(g, me.ID, "Viashino Sandstalker", "Creature — Lizard Warrior", viashinoSandstalkerOracle, true)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, lizard), "haste") {
		t.Error("Viashino Sandstalker has no haste")
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lizard) || !me.Hand.Contains(lizard) {
		t.Fatal("Viashino Sandstalker did not return to hand at your end step")
	}

	// On another player's turn it goes back at that turn's end.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToUpkeepOf(t, g2, 1)
	lizard2 := effectstest.PushCatalogPermanent(g2, me2.ID, "Viashino Sandstalker", "Creature — Lizard Warrior", viashinoSandstalkerOracle, true)
	effectstest.AdvanceToEndStepOf(t, g2, 1)
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(lizard2) || !me2.Hand.Contains(lizard2) {
		t.Fatal("Viashino Sandstalker did not return to hand at an opponent's end step")
	}
}

// --- Wicked Reward ----------------------------------------------------

// A creature is sacrificed to cast it; the target gets +4/+2.
func TestWickedRewardSacrificesAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := mm2Creature(g, me.ID, "Bear", 2, 2)
	victim := mm2Creature(g, me.ID, "Victim", 1, 1)
	id := effectstest.HandCardOf(g, me, "Wicked Reward", "Instant", "{1}{B}", wickedRewardOracle)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets}); err == nil {
		t.Fatal("cast with no sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets, SacrificeIDs: []uuid.UUID{victim}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the sacrifice was not paid while casting")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 6 || tough != 4 {
		t.Errorf("Bear is %d/%d, want 6/4", p, tough)
	}
}
