package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdr2S18SteelburrOracle = "d1f3d5ca-cb17-400f-8ed9-3b9a4f151b62"
	cmdr2S18BanditOracle    = "2a14670b-2ab6-4e1d-9b33-610a87bdb7ce"
)

func TestCmdr2S18SteelburrChampionGrowsOnOpponentNoncreatureSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champ := effectstest.B12Push(g, me.ID, "Steelburr Champion", "Creature — Mouse Soldier", cmdr2S18SteelburrOracle, 1, 1)
	effectstest.AssertKeywords(t, g, champ, "vigilance")
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFree(t, g, me, "My Bear", "Creature — Bear")
	effectstest.CastFree(t, g, me, "My Instant", "Instant")
	if got := effectstest.PlusOnes(g, champ); got != 0 {
		t.Fatalf("a creature spell or my own instant gave %d counters, want 0", got)
	}
	effectstest.CastFree(t, g, opp, "Their Instant", "Instant")
	if got := effectstest.PlusOnes(g, champ); got != 1 {
		t.Errorf("an opponent's instant left %d counters, want 1", got)
	}
}

func TestCmdr2S18SteelburrChampionOffspringMakesAOneOneCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbGLCastCreature(t, g, "Steelburr Champion", "Creature — Mouse Soldier", cmdr2S18SteelburrOracle, 1, 1, []int{0})
	for i := 0; i < 4; i++ {
		effectstest.BlbGLSettleTriggerOrder(g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CountNamed(g, "Steelburr Champion"); got != 2 {
		t.Fatalf("%d Steelburr Champions, want 2 (offspring paid)", got)
	}
}

func TestCmdr2S18ProsperousBanditMakesThatManyTappedTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bandit := effectstest.B12Push(g, me.ID, "Prosperous Bandit", "Creature — Raccoon Rogue", cmdr2S18BanditOracle, 2, 2)
	effectstest.AssertKeywords(t, g, bandit, "first strike")
	before := effectstest.CountNamed(g, "Treasure")
	effectstest.DealCombatDamageToPlayer(g, bandit, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountNamed(g, "Treasure") - before; got != 3 {
		t.Fatalf("created %d Treasures, want 3", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && !c.Tapped {
			t.Errorf("a Treasure entered untapped")
		}
	}
}

const cmdr2S18KillerServiceOracle = "0f30f32f-3c4c-4f73-8b2a-82045a40bda9"

func TestCmdr2S18KillerServiceMakesAFoodPerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Killer Service", "Enchantment", "", cmdr2S18KillerServiceOracle, game.CastSpellParams{FromZone: "hand"})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountNamed(g, "Food"); got != 3 {
		t.Errorf("created %d Foods, want 3 (one per opponent)", got)
	}
}

func TestCmdr2S18KillerServicePaysAndSacrificesATokenForARhino(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Killer Service", "Enchantment", cmdr2S18KillerServiceOracle, 0, 0)
	tok := effects.FoodToken()
	tok.InstanceID, tok.Owner, tok.Controller = uuid.New(), me.ID, me.ID
	food := effectstest.PushBattlefieldCardWithTimestamp(g, tok)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the end step offers the optional {2}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.AnswerChooseCardsOrPick(t, g, me.ID, food)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(food) {
		t.Error("the token was not sacrificed")
	}
	if got := effectstest.CountNamed(g, "Rhino Warrior"); got != 1 {
		t.Errorf("%d Rhino Warriors, want 1", got)
	}
}

const (
	cmdr2S18FlubsOracle     = "29356966-652d-4456-84c8-7aeab5fc8d05"
	cmdr2S18BumbleOracle    = "25636214-c6ea-4655-a5e8-98e735da5b05"
	cmdr2S18FoxgloveOracle  = "c2008ba9-00df-4607-ba0c-189af52033eb"
	cmdr2S18BelloOracle     = "fb773114-6161-46a3-ab30-1853958fbcb0"
	cmdr2S18HawkOracle      = "35a03b0f-9dac-476a-97e9-67767c1e453e"
	cmdr2S18FrugivoreOracle = "4f31eea0-61b9-467d-8393-0ab9f1b6e882"
	cmdr2S18TalentOracle    = "e802bf11-d3c7-4aad-8e85-99b39ff50d70"
	cmdr2S18ArthurOracle    = "8d6ce07f-b21b-42eb-a3d6-df810664e2be"
	cmdr2S18ArthurTypeLine  = "Legendary Creature — Mouse Knight"
	cmdr2S18FrugivoreTypeLn = "Creature — Rat Berserker"
)

func TestCmdr2S18EveryCardIsRegisteredFull(t *testing.T) {
	for _, id := range []string{cmdr2S18SteelburrOracle, cmdr2S18BanditOracle, cmdr2S18KillerServiceOracle,
		cmdr2S18FlubsOracle, cmdr2S18BumbleOracle, cmdr2S18FoxgloveOracle, cmdr2S18BelloOracle,
		cmdr2S18HawkOracle, cmdr2S18FrugivoreOracle, cmdr2S18TalentOracle, cmdr2S18ArthurOracle} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v %v", id, spec.Name, ok)
		}
	}
	if spec, _ := effects.Lookup(cmdr2S18FlubsOracle); spec.AdditionalLandPlays != 1 {
		t.Errorf("Flubs allows %d additional land plays, want 1", spec.AdditionalLandPlays)
	}
}

func TestCmdr2S18FlubsDiscardsWithCardsInHandAndDrawsWithNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Flubs, the Fool", "Legendary Creature — Frog Scout", cmdr2S18FlubsOracle, 0, 5)
	effectstest.AdvanceToMain(t, g)
	if me.Hand.Size() == 0 {
		t.Fatal("the test needs cards in hand")
	}
	effectstest.CastFree(t, g, me, "Shock", "Instant")
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("a spell cast with cards in hand should ask for a discard")
	}
	card := me.Hand.Cards[0].InstanceID
	effectstest.AnswerDiscard(t, g, me.ID, card)
	if me.Hand.Contains(card) {
		t.Error("the discarded card is still in hand")
	}

	g.WithWriteLock(func() { me.Hand.Cards = nil })
	lib := me.Library.Size()
	effectstest.CastFree(t, g, me, "Shock", "Instant")
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Fatal("an empty hand should draw, not discard")
	}
	if me.Hand.Size() != 1 || me.Library.Size() != lib-1 {
		t.Errorf("hand %d library %d (was %d): want one card drawn", me.Hand.Size(), me.Library.Size(), lib)
	}
}

func TestCmdr2S18MsBumbleflowerDrawsCountersAndDoublesOnTheSecondResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Ms. Bumbleflower", "Legendary Creature — Rabbit Citizen", cmdr2S18BumbleOracle, 1, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	oppHand := opp.Hand.Size()
	effectstest.CastFromHandForTest(t, g, me, "Shock", "Instant", "", "", game.CastSpellParams{FromZone: "hand"})
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - oppHand; got != 1 {
		t.Errorf("the opponent drew %d, want 1", got)
	}
	if got := effectstest.PlusOnes(g, bear); got != 1 {
		t.Errorf("the bear has %d counters, want 1", got)
	}
	effectstest.AssertKeywords(t, g, bear, "flying")
	mine := me.Hand.Size()
	effectstest.CastFromHandForTest(t, g, me, "Shock", "Instant", "", "", game.CastSpellParams{FromZone: "hand"})
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - mine; got != 2 {
		t.Errorf("the second resolution drew %d cards, want 2", got)
	}
}

func TestCmdr2S18MrFoxglovedrawsTheHandGap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fox := effectstest.B12Push(g, me.ID, "Mr. Foxglove", "Legendary Creature — Fox Rogue", cmdr2S18FoxgloveOracle, 3, 5)
	effectstest.AssertKeywords(t, g, fox, "lifelink")
	g.WithWriteLock(func() { me.Hand.Cards = me.Hand.Cards[:2] })
	oppHand := opp.Hand.Size()
	if oppHand <= 2 {
		t.Fatal("the test needs the defender to hold more cards")
	}
	effectstest.DeclareAttack(t, g, opp.ID, fox)
	effectstest.PassPriorityAroundTable(t, g)
	if want := 2 + (oppHand - 2); me.Hand.Size() != want {
		t.Errorf("hand %d, want %d", me.Hand.Size(), want)
	}
}

func TestCmdr2S18BelloMakesBigArtifactsAndEnchantmentsCreaturesOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Bello, Bard of the Brambles", "Legendary Creature — Raccoon Bard", cmdr2S18BelloOracle, 3, 3)
	push := func(name, typeLine, cost string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID,
		})
	}
	relic := push("Big Relic", "Artifact", "{4}")
	shrine := push("Big Shrine", "Enchantment", "{2}{G}{G}")
	small := push("Small Relic", "Artifact", "{3}")
	sword := push("Big Sword", "Artifact — Equipment", "{5}")
	aura := push("Big Aura", "Enchantment — Aura", "{4}")
	for _, id := range []uuid.UUID{relic, shrine} {
		if p, tough := effectstest.PtOf(t, g, id); p != 4 || tough != 4 {
			t.Errorf("%v is %d/%d, want 4/4", id, p, tough)
		}
		effectstest.AssertKeywords(t, g, id, "indestructible", "haste")
	}
	for _, id := range []uuid.UUID{small, sword, aura} {
		if p, _ := effectstest.PtOf(t, g, id); p == 4 {
			t.Errorf("%v was animated", id)
		}
	}
}

func TestCmdr2S18PyreswipeHawkGrowsWithArtifactsAndStealsOnExpendSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hawk := effectstest.B12Push(g, me.ID, "Pyreswipe Hawk", "Creature — Elemental Bird", cmdr2S18HawkOracle, 4, 4)
	effectstest.AssertKeywords(t, g, hawk, "flying", "haste")
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Relic", TypeLine: "Artifact", ManaCost: "{3}", Owner: me.ID, Controller: me.ID,
	})
	effectstest.DeclareAttack(t, g, opp.ID, hawk)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, hawk); got != 7 {
		t.Errorf("the attacking Hawk has power %d, want 7", got)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.B12Push(g2, me2.ID, "Pyreswipe Hawk", "Creature — Elemental Bird", cmdr2S18HawkOracle, 4, 4)
	theirs := effectstest.B12Permanent(g2, opp2.ID, "Their Relic", "Artifact")
	effectstest.ToMainForCost(t, g2)
	effectstest.CastPaid(t, g2, me2, "Six", "{6}", "CCCCCC")
	effectstest.AnswerPickTarget(t, g2, theirs)
	effectstest.PassPriorityAroundTable(t, g2)
	for _, c := range g2.Battlefield.Cards {
		if c.InstanceID == theirs && c.Controller != me2.ID {
			t.Error("the artifact was not taken")
		}
	}
}

func TestCmdr2S18InsatiableFrugivoreMakesAFoodAndSacrificesFoodsForAPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	effectstest.CastHoldCreature(t, g, "Insatiable Frugivore", cmdr2S18FrugivoreTypeLn, cmdr2S18FrugivoreOracle, 2, 4)
	if got := effectstest.CountNamed(g, "Food"); got != 1 {
		t.Fatalf("%d Foods, want 1 (no graveyard to repeat with)", got)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceTo(t, g2, game.StepPrecombatMain)
	frug := effectstest.PushCatalogPermanent(g2, me2.ID, "Insatiable Frugivore", cmdr2S18FrugivoreTypeLn, cmdr2S18FrugivoreOracle, false)
	bear := effectstest.B12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	mkFood := func() uuid.UUID {
		tok := effects.FoodToken()
		tok.InstanceID, tok.Owner, tok.Controller = uuid.New(), me2.ID, me2.ID
		return effectstest.PushBattlefieldCardWithTimestamp(g2, tok)
	}
	f1, f2 := mkFood(), mkFood()
	effectstest.FloatForTest(g2, me2, "BCCC")
	if err := g2.ActivateCatalogAbility(me2.ID, frug, 0, game.ActivateAbilityParams{
		XValue: 2, SacrificeIDs: []uuid.UUID{f1, f2},
	}); err != nil {
		t.Fatalf("activate at X=2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(f1) || g2.Battlefield.Contains(f2) {
		t.Error("the Foods were not sacrificed")
	}
	if got := effectstest.EffectivePower(t, g2, bear); got != 4 {
		t.Errorf("the bear has power %d, want 4", got)
	}
	effectstest.AssertKeywords(t, g2, bear, "menace")
}

func TestCmdr2S18GourmandsTalentLevelTwoMakesARaccoonOnTheFirstLifeGainOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	talent := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gourmand's Talent", TypeLine: "Enchantment — Class", OracleID: cmdr2S18TalentOracle,
		Owner: me.ID, Controller: me.ID, ClassLevel: 2,
	})
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(talent, me.ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(talent, me.ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountNamed(g, "Raccoon"); got != 1 {
		t.Errorf("%d Raccoons, want 1 (first life gain only)", got)
	}
}

func TestCmdr2S18GourmandsTalentLevelThreeCountersEachCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	talent := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gourmand's Talent", TypeLine: "Enchantment — Class", OracleID: cmdr2S18TalentOracle,
		Owner: me.ID, Controller: me.ID, ClassLevel: 3,
	})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(talent, me.ID, 3) })
	// Level 3 keeps level 2's line, so two triggers go on the stack.
	for i := 0; i < 4; i++ {
		effectstest.BlbGLSettleTriggerOrder(g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.PlusOnes(g, bear); got < 1 {
		t.Errorf("the bear has %d counters, want at least 1", got)
	}
	if got := effectstest.CountNamed(g, "Raccoon"); got != 1 {
		t.Errorf("%d Raccoons, want 1 (level 2's line still applies)", got)
	}
}

func TestCmdr2S18GourmandsTalentMakesYourArtifactsFoodsOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gourmand's Talent", TypeLine: "Enchantment — Class", OracleID: cmdr2S18TalentOracle,
		Owner: me.ID, Controller: me.ID, ClassLevel: 1,
	})
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == relic && !c.HasSubtype("Food") {
				t.Error("the artifact is not a Food on its controller's turn")
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CC")
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, relic, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the granted Food ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact was not sacrificed")
	}
}

func TestCmdr2S18ArthurPutsACreatureAttackingAndReturnsItAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arthur := effectstest.B12Push(g, me.ID, "Arthur, Marigold Knight", cmdr2S18ArthurTypeLine, cmdr2S18ArthurOracle, 4, 5)
	friend := effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 1, 1)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Dragon", TypeLine: "Creature — Dragon", Power: 5, Toughness: 5},
		game.Card{Name: "Filler 1", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 2", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 3", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 4", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 5", TypeLine: "Sorcery"},
		game.Card{Name: "Filler 6", TypeLine: "Sorcery"})
	effectstest.DeclareAttack(t, g, opp.ID, arthur, friend)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0])
	effectstest.FinB12AnswerAttackPick(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[0]) {
		t.Fatal("the Dragon did not enter the battlefield")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ids[0]) {
		t.Error("the Dragon was not returned at end of combat")
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the Dragon is not in its owner's hand")
	}
}

func TestCmdr2S18ArthurNeedsAnotherAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arthur := effectstest.B12Push(g, me.ID, "Arthur, Marigold Knight", cmdr2S18ArthurTypeLine, cmdr2S18ArthurOracle, 4, 5)
	lib := me.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, arthur)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != lib {
		t.Error("Arthur attacking alone looked at the library")
	}
}
