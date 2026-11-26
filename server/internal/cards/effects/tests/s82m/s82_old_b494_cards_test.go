package testss82m

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b494_cards_test.go — Zendikar enters triggers, S82 OLD batch 494
// (#5583), each driven through the engine.

const (
	zenBalaGed   = "3a020ae9-279b-4c6c-8d3d-90d7b335fa3c"
	zenWayfinder = "45b34d33-7d70-4308-8f07-bb044be891bb"
	zenTiger     = "4d0fa514-87fc-46ff-bf49-4cd332396978"
	zenEvangel   = "83b480ea-77c5-4e00-99d0-bea2a60a1bfe"
	zenOutfitter = "8ae3562f-28b7-4462-96ed-be0cf7052ccc"
	zenKhalniGem = "b0a96e47-91da-43ee-93dd-f866721029c6"
)

func TestS82OldB494EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		zenBalaGed: "Bala Ged Thief", zenWayfinder: "Merfolk Wayfinder", zenTiger: "Scythe Tiger",
		zenEvangel: "Kabira Evangel", zenOutfitter: "Kor Outfitter", zenKhalniGem: "Khalni Gem",
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

func zenCast(t *testing.T, g *game.Game, c game.Card) uuid.UUID {
	t.Helper()
	id := effectstest.TdCastFromHand(t, g, c, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestS82OldB494KorOutfitterAttachesTheChosenEquipmentOnAYes(t *testing.T) {
	for _, yes := range []bool{true, false} {
		t.Run(fmt.Sprintf("yes=%v", yes), func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			bear := effectstest.SeedBear(g, me.ID)
			sword := effectstest.SeedEquipment(g, me.ID, "Test Sword", "")
			zenCast(t, g, game.Card{Name: "Kor Outfitter", TypeLine: "Creature — Kor Soldier", OracleID: zenOutfitter, Power: 2, Toughness: 2})
			effectstest.B04WaitForPick(t, g, me.ID)
			effectstest.PickCard(t, g, me.ID, sword)
			effectstest.B04WaitForPick(t, g, me.ID)
			effectstest.PickCard(t, g, me.ID, bear)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
			effectstest.PassPriorityAroundTable(t, g)
			attached := effectstest.AttachmentHostOf(t, g, sword).ID == bear
			if attached != yes {
				t.Errorf("attached = %v, want %v", attached, yes)
			}
		})
	}
}

func TestS82OldB494BalaGedThiefRevealsAsManyCardsAsYouHaveAlliesAndYouPickOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "Other Ally", "Creature — Human Ally", 1, 1)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	var hand []uuid.UUID
	for i := 1; i <= 4; i++ {
		id := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: fmt.Sprintf("Their Card %d", i), TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
		hand = append(hand, id)
	}
	// The Thief enters beside one other Ally: two Allies, two cards.
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Bala Ged Thief", TypeLine: "Creature — Human Rogue Ally", OracleID: zenBalaGed, Power: 2, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if ask == nil || ask.ChooseMin != 2 || ask.ChooseMax != 2 {
		t.Fatalf("the target should reveal exactly two cards, got %+v", ask)
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, hand[0], hand[1])
	rp := effectstest.RevealPickFor(g, me.ID)
	if rp == nil || len(rp.ChooseCards) != 2 {
		t.Fatalf("want two revealed cards to pick from, got %+v", rp)
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{hand[1]}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(hand[1]) || !opp.Graveyard.Contains(hand[1]) {
		t.Error("the chosen card was not discarded")
	}
	if !opp.Hand.Contains(hand[0]) || len(opp.Hand.Cards) != 3 {
		t.Errorf("the other revealed card or the unrevealed ones were touched (hand %d)", len(opp.Hand.Cards))
	}
}

func TestS82OldB494BalaGedThiefTriggersForAnotherAllyEnteringToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Bala Ged Thief", "Creature — Human Rogue Ally", zenBalaGed, 2, 2)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	a := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: a, Name: "Lone Card", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	// A non-Ally creature does not trigger it.
	zenCast(t, g, game.Card{Name: "Plain Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a non-Ally entering triggered the Thief")
	}
	// An Ally does; two Allies, but the hand holds one card, so it is all revealed.
	effectstest.TdCastFromHand(t, g, game.Card{Name: "New Ally", TypeLine: "Creature — Kor Ally", Power: 1, Toughness: 1}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	rp := effectstest.RevealPickFor(g, me.ID)
	if rp == nil {
		t.Fatal("no reveal pick after another Ally entered")
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{a}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(a) {
		t.Error("the card was not discarded")
	}
}

func TestS82OldB494MerfolkWayfinderTakesTheIslandsAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	push := func(name, typeLine string) uuid.UUID {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		return id
	}
	fourth := push("Fourth", "Sorcery")
	island2 := push("Island", "Basic Land — Island")
	forest := push("Forest", "Basic Land — Forest")
	island1 := push("Island", "Basic Land — Island")
	hand := me.Hand.Size()
	zenCast(t, g, game.Card{Name: "Merfolk Wayfinder", TypeLine: "Creature — Merfolk Scout", OracleID: zenWayfinder, Power: 1, Toughness: 2, Keywords: []string{"flying"}})
	if !me.Hand.Contains(island1) || !me.Hand.Contains(island2) {
		t.Error("an Island was not put into hand")
	}
	if me.Hand.Contains(forest) || me.Hand.Size() != hand+2 {
		t.Errorf("hand %d -> %d, want the Wayfinder out and two Islands in", hand, me.Hand.Size())
	}
	if !me.Library.Contains(forest) || !me.Library.Contains(fourth) {
		t.Fatal("the Forest or the fourth card left the library")
	}
	if me.Library.Cards[0].InstanceID != forest {
		t.Error("the Forest is not on the bottom of the library")
	}
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID != fourth {
		t.Error("the fourth card is not the new top")
	}
}

func TestS82OldB494ScytheTigerSacrificesALandOrItself(t *testing.T) {
	t.Run("a land", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		land := effectstest.PushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
		tiger := zenCast(t, g, game.Card{Name: "Scythe Tiger", TypeLine: "Creature — Cat", OracleID: zenTiger, Power: 3, Toughness: 2})
		effectstest.AnswerOptionPick(t, g, me.ID, 0)
		effectstest.AnswerChooseCards(t, g, me.ID, land)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(land) || !g.Battlefield.Contains(tiger) {
			t.Error("the land should be sacrificed and the Tiger kept")
		}
	})
	t.Run("no land", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		tiger := zenCast(t, g, game.Card{Name: "Scythe Tiger", TypeLine: "Creature — Cat", OracleID: zenTiger, Power: 3, Toughness: 2})
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(tiger) {
			t.Error("the Tiger stayed with no land to sacrifice")
		}
	})
	t.Run("declined", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		land := effectstest.PushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
		tiger := zenCast(t, g, game.Card{Name: "Scythe Tiger", TypeLine: "Creature — Cat", OracleID: zenTiger, Power: 3, Toughness: 2})
		effectstest.AnswerOptionPick(t, g, me.ID, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(tiger) || !g.Battlefield.Contains(land) {
			t.Error("declining should sacrifice the Tiger and keep the land")
		}
	})
}

func TestS82OldB494ScytheTigerHasShroud(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tiger := effectstest.PushCatalogPermanent(g, me.ID, "Scythe Tiger", "Creature — Cat", zenTiger, false)
	if !effectstest.P1g2HasKeyword(g, tiger, "shroud") {
		t.Error("Scythe Tiger lacks shroud")
	}
}

func TestS82OldB494KabiraEvangelGivesAlliesProtectionFromTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ally := effectstest.B12Creature(g, me.ID, "Other Ally", "Creature — Human Ally", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	evangel := effectstest.TdCastFromHand(t, g, game.Card{Name: "Kabira Evangel", TypeLine: "Creature — Human Cleric Ally", OracleID: zenEvangel, Power: 2, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	red := game.ProtectionFromColor("R")
	if !effectstest.P1g2HasKeyword(g, ally, red) || !effectstest.P1g2HasKeyword(g, evangel, red) {
		t.Error("an Ally did not gain protection from red")
	}
	if effectstest.P1g2HasKeyword(g, bear, red) {
		t.Error("a non-Ally gained protection from red")
	}
	// It wears off at end of turn.
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	effectstest.Settle(t, g)
	if effectstest.P1g2HasKeyword(g, ally, red) {
		t.Error("the protection outlasted the turn")
	}
}

func TestS82OldB494KabiraEvangelDeclinedAsksNoColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	evangel := effectstest.TdCastFromHand(t, g, game.Card{Name: "Kabira Evangel", TypeLine: "Creature — Human Cleric Ally", OracleID: zenEvangel, Power: 2, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.P1g2HasKeyword(g, evangel, game.ProtectionFromColor("R")) {
		t.Error("protection granted after declining")
	}
}

func TestS82OldB494KhalniGemReturnsTwoLandsAndTapsForTwoOfOneColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	var lands []uuid.UUID
	for i := 0; i < 3; i++ {
		lands = append(lands, effectstest.PushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest"))
	}
	gem := zenCast(t, g, game.Card{Name: "Khalni Gem", TypeLine: "Artifact", OracleID: zenKhalniGem})
	effectstest.AnswerOwnPermanents(t, g, me.ID, lands[0], lands[2])
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lands[0]) || g.Battlefield.Contains(lands[2]) || !g.Battlefield.Contains(lands[1]) {
		t.Error("the two chosen lands should be returned and the other kept")
	}
	if !me.Hand.Contains(lands[0]) || !me.Hand.Contains(lands[2]) {
		t.Error("the lands are not in their owner's hand")
	}
	if !g.Battlefield.Contains(gem) {
		t.Fatal("the Gem left the battlefield")
	}
	if err := g.ActivateManaAbility(me.ID, gem, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || len(pick.ColorOptions) != 5 || pick.ManaAmounts["W"] != 2 {
		t.Fatalf("pick = %+v, want one five-colour pick of two", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("pool = %v, want two {G}", got)
	}
}

func TestS82OldB494KhalniGemWithOneLandReturnsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	zenCast(t, g, game.Card{Name: "Khalni Gem", TypeLine: "Artifact", OracleID: zenKhalniGem})
	if g.Battlefield.Contains(land) {
		effectstest.AnswerOwnPermanents(t, g, me.ID, land)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(land) || !me.Hand.Contains(land) {
		t.Error("the only land should be returned")
	}
}
