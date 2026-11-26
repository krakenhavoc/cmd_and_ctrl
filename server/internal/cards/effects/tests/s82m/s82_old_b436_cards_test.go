package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 set AVR, batch 436 (#5570): Avacyn Restored, soulbond, creatures with
// keywords, card draw and library.

const (
	avr436Stonewright   = "2fe2e0c4-6e02-4f5a-b2b2-920bf912a2bd"
	avr436Alchemist     = "694b78c6-443a-4044-84ea-46ac85c8d2cd"
	avr436SternMentor   = "6dab0e1a-f39a-4113-9d13-b067cfefbdb3"
	avr436DiregrafEscrt = "f28a67d8-05ec-4d72-aa17-670ce1136f5c"
	avr436Fettergeist   = "7ed73cff-137c-4ad1-b11d-c62aa746655e"
	avr436ArchwingDrgn  = "a0a706b2-b237-4183-82fa-2140a23b89e3"
	avr436MisthollowGrf = "a990e744-bfc4-435d-aef6-6256ac10f5a6"
	avr436MidnightDuel  = "e8e2e1e9-13a6-4cfd-b9ed-2080e984844d"
	avr436AmassComps    = "0f623e23-81ae-4c35-b68b-e0a2f431cbad"
	avr436DangerousWgr  = "7e0d60d7-cb27-4e45-80d4-c29aef5d56b6"
	avr436MassAppeal    = "aef067b4-cd9c-4f18-8c63-fbbc39b63376"
)

func avr436Registered(t *testing.T) {
	t.Helper()
	for oracle, name := range map[string]string{
		avr436Stonewright: "Stonewright", avr436Alchemist: "Galvanic Alchemist", avr436SternMentor: "Stern Mentor",
		avr436DiregrafEscrt: "Diregraf Escort", avr436Fettergeist: "Fettergeist", avr436ArchwingDrgn: "Archwing Dragon",
		avr436MisthollowGrf: "Misthollow Griffin", avr436MidnightDuel: "Midnight Duelist",
		avr436AmassComps: "Amass the Components", avr436DangerousWgr: "Dangerous Wager", avr436MassAppeal: "Mass Appeal",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func TestAvrB436RegisteredFull(t *testing.T) { avr436Registered(t) }

// avr436Pair puts a soulbond creature and a Bear on the battlefield, paired.
func avr436Pair(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int) (soul, bear uuid.UUID) {
	t.Helper()
	soul = effectstest.B12Push(g, owner, name, typeLine, oracle, p, tough)
	bear = effectstest.B12Push(g, owner, "Bear", effectstest.TestCreatureTypeLine, "", 2, 2)
	g.WithWriteLock(func() {
		if !g.PairForEffect(owner, soul, bear) {
			t.Fatal("could not pair the creatures")
		}
	})
	return soul, bear
}

func TestAvrB436StonewrightGivesBothCreaturesAFirebreathingAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	soul, bear := avr436Pair(t, g, me.ID, "Stonewright", "Creature — Human Shaman", avr436Stonewright, 1, 1)
	for _, id := range []uuid.UUID{soul, bear} {
		effectstest.B06AddMana(me, "R")
		before, _ := effectstest.EffectivePTOf(t, g, id)
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate the granted ability: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if after, _ := effectstest.EffectivePTOf(t, g, id); after != before+1 {
			t.Errorf("power %d, want %d", after, before+1)
		}
	}
}

func TestAvrB436StonewrightGrantsNothingUnpaired(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Stonewright", "Creature — Human Shaman", avr436Stonewright, 1, 1)
	bear := effectstest.B12Push(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, "", 2, 2)
	effectstest.B06AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("an unpaired creature has no granted ability")
	}
}

func TestAvrB436GalvanicAlchemistUntapsEitherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_, bear := avr436Pair(t, g, me.ID, "Galvanic Alchemist", "Creature — Human Wizard", avr436Alchemist, 1, 4)
	if err := g.TapCard(bear, true); err != nil {
		t.Fatal(err)
	}
	effectstest.B06AddMana(me, "U", "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, bear) {
		t.Error("the partner should be untapped")
	}
}

func TestAvrB436SternMentorMillsTwoWithTheTapAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, bear := avr436Pair(t, g, me.ID, "Stern Mentor", "Creature — Human Wizard", avr436SternMentor, 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).SummonedThisTurn = false })
	lib := opp.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.AcPlayerRef(opp.ID)}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != lib-2 {
		t.Errorf("library = %d, want %d", opp.Library.Size(), lib-2)
	}
}

func TestAvrB436DiregrafEscortGivesBothProtectionFromZombiesWhilePaired(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	soul := effectstest.B12Push(g, me.ID, "Diregraf Escort", "Creature — Human Cleric", avr436DiregrafEscrt, 1, 1)
	bear := effectstest.B12Push(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, "", 2, 2)
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "protection from Zombies") {
		t.Fatal("an unpaired creature should have no protection")
	}
	g.WithWriteLock(func() { g.PairForEffect(me.ID, soul, bear) })
	for _, id := range []uuid.UUID{soul, bear} {
		if !effectstest.EffectiveAbilitiesContain(t, g, id, "protection from Zombies") {
			t.Error("a paired creature should have protection from Zombies")
		}
	}
}

func TestAvrB436MidnightDuelistHasProtectionFromVampires(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Midnight Duelist", "Creature — Human Soldier", avr436MidnightDuel, 1, 2)
	effectstest.AssertKeywords(t, g, id, "protection from Vampires")
}

func TestAvrB436ArchwingDragonFliesHastesAndReturnsAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Archwing Dragon", "Creature — Dragon", avr436ArchwingDrgn, 4, 4)
	effectstest.AssertKeywords(t, g, id, "flying", "haste")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("the Dragon should have returned to its owner's hand")
	}
	found := false
	for _, c := range me.Hand.Cards {
		if c.Name == "Archwing Dragon" {
			found = true
		}
	}
	if !found {
		t.Error("the Dragon is not in its owner's hand")
	}
}

func TestAvrB436MisthollowGriffinCanBeCastFromExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	griffin := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: griffin, Name: "Misthollow Griffin", TypeLine: "Creature — Griffin", ManaCost: "{2}{U}{U}",
			OracleID: avr436MisthollowGrf, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	})
	effectstest.FillPool(me, 3)
	if err := g.CastSpell(me.ID, griffin, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Fatal("a four-mana spell was cast with three mana")
	}
	effectstest.B06AddMana(me, "U", "U", "U", "U")
	me.ManaPool.EmptyPool()
	effectstest.B06AddMana(me, "U", "U", "U", "U")
	if err := g.CastSpell(me.ID, griffin, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast the Griffin from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(griffin) {
		t.Error("the Griffin cast from exile did not resolve onto the battlefield")
	}
}

func avr436Upkeep(t *testing.T) (g *game.Game, me *game.Player, geist uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	geist = effectstest.B12Push(g, me.ID, "Fettergeist", "Creature — Spirit", avr436Fettergeist, 3, 4)
	effectstest.B12Push(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, "", 2, 2)
	effectstest.B12Push(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, "", 2, 2)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	return g, me, geist
}

func TestAvrB436FettergeistStaysWhenYouPayOnePerOtherCreature(t *testing.T) {
	g, me, geist := avr436Upkeep(t)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no pay prompt for the upkeep trigger")
	}
	c := effectstest.AnswerPayUnless(t, g, me.ID, true)
	if c == nil {
		t.Fatal("no prompt answered")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(geist) {
		t.Error("the Fettergeist should stay after paying {2}")
	}
}

func TestAvrB436FettergeistIsSacrificedWhenYouDontPay(t *testing.T) {
	g, me, geist := avr436Upkeep(t)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		effectstest.AnswerPayUnless(t, g, me.ID, false)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(geist) {
		t.Error("the Fettergeist should be sacrificed")
	}
}

func TestAvrB436FettergeistAloneCostsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	geist := effectstest.B12Push(g, me.ID, "Fettergeist", "Creature — Spirit", avr436Fettergeist, 3, 4)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("no payment is owed with no other creature")
	}
	if !g.Battlefield.Contains(geist) {
		t.Error("the lone Fettergeist stays")
	}
}

func TestAvrB436AmassTheComponentsDrawsThreeThenBottomsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 5; i++ {
		effectstest.SeedLibraryTop(me, "Filler", "Sorcery", "")
	}
	hand, lib := me.Hand.Size(), me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Amass the Components", "Sorcery", avr436AmassComps, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCards(g, me.ID)
	if c == nil {
		t.Fatal("no hand pick for the bottom card")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, c.ChooseCards[:1]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand = %d, want %d (draw three, put one back)", me.Hand.Size(), hand+2)
	}
	if me.Library.Size() != lib-2 {
		t.Errorf("library = %d, want %d", me.Library.Size(), lib-2)
	}
}

func TestAvrB436DangerousWagerDiscardsTheHandThenDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		effectstest.PushHandCard(g, me)
		effectstest.SeedLibraryTop(me, "Filler", "Sorcery", "")
	}
	gy := me.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Dangerous Wager", "Instant", avr436DangerousWgr, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 2 {
		t.Errorf("hand = %d, want 2 (it was %d before the spell resolved)", me.Hand.Size(), hand)
	}
	if me.Graveyard.Size() < gy+hand+1 {
		t.Errorf("graveyard = %d, want at least %d", me.Graveyard.Size(), gy+hand+1)
	}
}

func TestAvrB436MassAppealDrawsPerHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 6; i++ {
		effectstest.SeedLibraryTop(me, "Filler", "Sorcery", "")
	}
	effectstest.B12Push(g, me.ID, "Human A", "Creature — Human Soldier", "", 1, 1)
	effectstest.B12Push(g, me.ID, "Human B", "Creature — Human Wizard", "", 1, 1)
	effectstest.B12Push(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, "", 2, 2)
	effectstest.B12Push(g, opp.ID, "Their Human", "Creature — Human Soldier", "", 1, 1)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Mass Appeal", "Sorcery", avr436MassAppeal, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand = %d, want %d (two Humans you control)", me.Hand.Size(), hand+2)
	}
}
