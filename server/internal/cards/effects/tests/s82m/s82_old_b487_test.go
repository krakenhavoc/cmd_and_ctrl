package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 487 (#5582): Worldwake, activated abilities and counters.

const (
	wwk487VastwoodAnimist   = "0b331f27-0d78-4d22-b546-36c5dec37606"
	wwk487BazaarTrader      = "21ad7067-3210-42d6-9009-0784b2834227"
	wwk487WalkingAtlas      = "855a4837-7fc8-4b97-afbd-daa88e322c89"
	wwk487AgadeemOccultist  = "9c462cd8-36f9-4e39-adf3-310ad4d76051"
	wwk487HarabazDruid      = "ead985ec-f29f-4a3b-b8b1-061142cc5bd1"
	wwk487QuestGoblinLord   = "2e19baba-2d66-4e3a-b2bf-cb27dce89f44"
	wwk487QuestUlasTemple   = "434807c5-5948-4181-a46c-2c834d0a2f0b"
	wwk487QuestNihilStone   = "7f65178e-dd2e-44aa-a68f-f8694e0b250a"
	wwk487TalusPaladin      = "98914c0e-967f-4c8b-bd74-e2ce15c7fe95"
	wwk487FeralContest      = "efbb5b87-d175-477e-a490-85dd98b62de6"
	wwk487AllyType          = "Creature — Human Ally"
	wwk487BearType          = "Creature — Bear"
	wwk487ForestType        = "Basic Land — Forest"
	wwk487EnchantmentTypeLn = "Enchantment"
)

func TestWwkB487CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		wwk487VastwoodAnimist: "Vastwood Animist", wwk487BazaarTrader: "Bazaar Trader",
		wwk487WalkingAtlas: "Walking Atlas", wwk487AgadeemOccultist: "Agadeem Occultist",
		wwk487HarabazDruid: "Harabaz Druid", wwk487QuestGoblinLord: "Quest for the Goblin Lord",
		wwk487QuestUlasTemple: "Quest for Ula's Temple", wwk487QuestNihilStone: "Quest for the Nihil Stone",
		wwk487TalusPaladin: "Talus Paladin", wwk487FeralContest: "Feral Contest",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

// --- Vastwood Animist --------------------------------------------------------

func TestWwkB487VastwoodAnimistMakesALandAnXByXForTheAlliesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	animist := effectstest.B12Push(g, me.ID, "Vastwood Animist", "Creature — Elf Shaman Ally", wwk487VastwoodAnimist, 1, 1)
	effectstest.B12Creature(g, me.ID, "Other Ally", wwk487AllyType, 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", wwk487BearType, 2, 2)
	land := effectstest.FreeSpellPermanent(g, me.ID, "Forest", wwk487ForestType)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, animist, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(land)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, land)
	if !c.IsCreature() || !c.IsLand() {
		t.Fatalf("the Forest is not a land creature: %v", c.Effective().Types)
	}
	if p, tough := effectstest.PtOf(t, g, land); p != 2 || tough != 2 {
		t.Errorf("the land is %d/%d, want 2/2 (two Allies)", p, tough)
	}
	effectstest.EndTurn(t, g)
	if effectstest.LayeredCard(t, g, land).IsCreature() {
		t.Error("the land is still a creature after the turn ended")
	}
}

func TestWwkB487VastwoodAnimistRefusesAnOpponentsLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	animist := effectstest.B12Push(g, me.ID, "Vastwood Animist", "Creature — Elf Shaman Ally", wwk487VastwoodAnimist, 1, 1)
	land := effectstest.FreeSpellPermanent(g, opp.ID, "Forest", wwk487ForestType)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, animist, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(land)}); err == nil {
		t.Fatal("an opponent's land was accepted")
	}
}

// --- Bazaar Trader -----------------------------------------------------------

func TestWwkB487BazaarTraderGivesAPermanentAwayForGood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trader := effectstest.B12Push(g, me.ID, "Bazaar Trader", "Creature — Goblin", wwk487BazaarTrader, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", wwk487BearType, 2, 2)
	effectstest.AdvanceToMain(t, g)
	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0}, {Kind: game.TargetCard, ID: bear, Slot: 1}}
	if err := g.ActivateCatalogAbility(me.ID, trader, 0, game.ActivateAbilityParams{Targets: targets}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, bear).Controller; got != opp.ID {
		t.Errorf("the Bear is controlled by %v, want the opponent", got)
	}
	effectstest.EndTurn(t, g)
	if got := effectstest.LayeredCard(t, g, bear).Controller; got != opp.ID {
		t.Errorf("the Bear went back at end of turn (controlled by %v)", got)
	}
}

func TestWwkB487BazaarTraderCanGiveAwayALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trader := effectstest.B12Push(g, me.ID, "Bazaar Trader", "Creature — Goblin", wwk487BazaarTrader, 1, 1)
	land := effectstest.FreeSpellPermanent(g, me.ID, "Forest", wwk487ForestType)
	effectstest.AdvanceToMain(t, g)
	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0}, {Kind: game.TargetCard, ID: land, Slot: 1}}
	if err := g.ActivateCatalogAbility(me.ID, trader, 0, game.ActivateAbilityParams{Targets: targets}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, land).Controller; got != opp.ID {
		t.Errorf("the Forest is controlled by %v, want the opponent", got)
	}
}

func TestWwkB487BazaarTraderRefusesAPermanentYouDoNotControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trader := effectstest.B12Push(g, me.ID, "Bazaar Trader", "Creature — Goblin", wwk487BazaarTrader, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", wwk487BearType, 2, 2)
	effectstest.AdvanceToMain(t, g)
	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1}}
	if err := g.ActivateCatalogAbility(me.ID, trader, 0, game.ActivateAbilityParams{Targets: targets}); err == nil {
		t.Fatal("a permanent you do not control was accepted")
	}
}

// --- Walking Atlas -----------------------------------------------------------

func TestWwkB487WalkingAtlasPutsALandFromHandOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	atlas := effectstest.B12Push(g, me.ID, "Walking Atlas", "Artifact Creature — Construct", wwk487WalkingAtlas, 1, 1)
	land := effectstest.HandCardForTest(me, "Forest", wwk487ForestType, "")
	bear := effectstest.HandCardForTest(me, "Bear", wwk487BearType, "")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, atlas, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) || effectstest.TappedOnBattlefield(t, g, land) {
		t.Error("the Forest should be on the battlefield, untapped")
	}
	if !me.Hand.Contains(bear) {
		t.Error("a nonland card left the hand")
	}
}

func TestWwkB487WalkingAtlasMayDeclineAndAsksNothingWithoutALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	atlas := effectstest.B12Push(g, me.ID, "Walking Atlas", "Artifact Creature — Construct", wwk487WalkingAtlas, 1, 1)
	land := effectstest.HandCardForTest(me, "Forest", wwk487ForestType, "")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, atlas, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if g.Battlefield.Contains(land) {
		t.Error("declining still put the land onto the battlefield")
	}
	// The second activation needs an untap; a hand with no land asks nothing.
	me.Hand.Cards = nil
	effectstest.UntapForTest(g, atlas)
	if err := g.ActivateCatalogAbility(me.ID, atlas, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate again: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a hand with no land was asked to choose one")
	}
}

// --- Agadeem Occultist -------------------------------------------------------

func wwk487Occultist(t *testing.T, cost string, allies int) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, error) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	occ := effectstest.B12Push(g, me.ID, "Agadeem Occultist", "Creature — Human Shaman Ally", wwk487AgadeemOccultist, 0, 2)
	for i := 1; i < allies; i++ {
		effectstest.B12Creature(g, me.ID, "Ally", wwk487AllyType, 1, 1)
	}
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: wwk487BearType, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	err := g.ActivateCatalogAbility(me.ID, occ, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(dead)})
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return g, me, dead, occ, err
}

func TestWwkB487AgadeemOccultistReanimatesACardWithinTheAllyCount(t *testing.T) {
	g, me, dead, _, err := wwk487Occultist(t, "{1}{G}", 2)
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature card (mana value 2, two Allies) was not put onto the battlefield")
	}
	if got := effectstest.LayeredCard(t, g, dead).Controller; got != me.ID {
		t.Errorf("it is controlled by %v, want you", got)
	}
}

func TestWwkB487AgadeemOccultistLeavesACardWithTooHighAManaValue(t *testing.T) {
	g, _, dead, _, err := wwk487Occultist(t, "{2}{G}", 2)
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(dead) {
		t.Error("a mana value 3 card came back with only two Allies")
	}
}

// --- Harabaz Druid -----------------------------------------------------------

func TestWwkB487HarabazDruidAddsOneManaPerAlly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	druid := effectstest.B12Push(g, me.ID, "Harabaz Druid", "Creature — Human Druid Ally", wwk487HarabazDruid, 0, 1)
	effectstest.B12Creature(g, me.ID, "Ally A", wwk487AllyType, 1, 1)
	effectstest.B12Creature(g, me.ID, "Ally B", wwk487AllyType, 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", wwk487BearType, 2, 2)
	effectstest.UntapForTest(g, druid)
	effectstest.ActivateManaFor(t, g, me.ID, druid, 0, game.ManaAbilityParams{})
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || pick.ManaAmounts["U"] != 3 {
		t.Fatalf("pick = %+v, want amounts of three (the Druid and two Allies)", pick)
	}
}

// --- Quest for the Goblin Lord -----------------------------------------------

func TestWwkB487QuestForTheGoblinLordCountsGoblinsAndPumpsAtFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	quest := effectstest.B12Push(g, me.ID, "Quest for the Goblin Lord", wwk487EnchantmentTypeLn, wwk487QuestGoblinLord, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", wwk487BearType, 2, 2)
	effectstest.CastHoldCreature(t, g, "Goblin", "Creature — Goblin", "", 1, 1)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B18Counters(t, g, quest, "quest"); n != 1 {
		t.Fatalf("quest counters %d, want 1", n)
	}
	effectstest.AddCountersForTest(g, quest, "quest", 3)
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Errorf("the Bear has power %d with four counters, want 2", p)
	}
	effectstest.AddCountersForTest(g, quest, "quest", 1)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the Bear is %d/%d with five counters, want 4/2", p, tough)
	}
}

func TestWwkB487QuestForTheGoblinLordIgnoresNonGoblinsAndMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	quest := effectstest.B12Push(g, me.ID, "Quest for the Goblin Lord", wwk487EnchantmentTypeLn, wwk487QuestGoblinLord, 0, 0)
	effectstest.CastHoldCreature(t, g, "Bear", wwk487BearType, "", 2, 2)
	if effectstest.B18TriggerPromptCount(g, me.ID) != 0 {
		t.Fatal("a non-Goblin entering asked about a quest counter")
	}
	effectstest.CastHoldCreature(t, g, "Goblin", "Creature — Goblin", "", 1, 1)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B18Counters(t, g, quest, "quest"); n != 0 {
		t.Errorf("quest counters %d after declining, want 0", n)
	}
}

// --- Quest for Ula's Temple --------------------------------------------------

func TestWwkB487QuestForUlasTempleCountsACreatureOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	quest := effectstest.B12Push(g, me.ID, "Quest for Ula's Temple", wwk487EnchantmentTypeLn, wwk487QuestUlasTemple, 0, 0)
	effectstest.SeedLibraryTop(me, "Sea Beast", "Creature — Serpent", "{3}{U}")
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B18Counters(t, g, quest, "quest"); n != 1 {
		t.Errorf("quest counters %d, want 1", n)
	}
}

func TestWwkB487QuestForUlasTempleLeavesANonCreatureOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	quest := effectstest.B12Push(g, me.ID, "Quest for Ula's Temple", wwk487EnchantmentTypeLn, wwk487QuestUlasTemple, 0, 0)
	effectstest.SeedLibraryTop(me, "Island", "Basic Land — Island", "")
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Fatal("a land on top offered to reveal it")
	}
	if n := effectstest.B18Counters(t, g, quest, "quest"); n != 0 {
		t.Errorf("quest counters %d, want 0", n)
	}
}

func TestWwkB487QuestForUlasTemplePutsASeaCreatureFromHandAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	quest := effectstest.B12Push(g, me.ID, "Quest for Ula's Temple", wwk487EnchantmentTypeLn, wwk487QuestUlasTemple, 0, 0)
	effectstest.AddCountersForTest(g, quest, "quest", 3)
	kraken := effectstest.HandCardForTest(me, "Kraken", "Creature — Kraken", "")
	bear := effectstest.HandCardForTest(me, "Bear", wwk487BearType, "")
	effectstest.AdvanceToStepInTurn(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, kraken)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(kraken) {
		t.Error("the Kraken was not put onto the battlefield")
	}
	if !me.Hand.Contains(bear) {
		t.Error("the Bear left the hand")
	}
}

func TestWwkB487QuestForUlasTempleNeedsThreeCountersAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	quest := effectstest.B12Push(g, me.ID, "Quest for Ula's Temple", wwk487EnchantmentTypeLn, wwk487QuestUlasTemple, 0, 0)
	effectstest.AddCountersForTest(g, quest, "quest", 2)
	kraken := effectstest.HandCardForTest(me, "Kraken", "Creature — Kraken", "")
	effectstest.AdvanceToStepInTurn(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil || g.Battlefield.Contains(kraken) {
		t.Error("two quest counters should not trigger the end-step put")
	}
}

// --- Quest for the Nihil Stone -----------------------------------------------

func TestWwkB487QuestForTheNihilStoneCountsOpponentDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	quest := effectstest.B12Push(g, me.ID, "Quest for the Nihil Stone", wwk487EnchantmentTypeLn, wwk487QuestNihilStone, 0, 0)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	if effectstest.B19HasTriggerPromptFor(g, me.ID) {
		t.Error("your own discard is not an opponent's")
	}
	if n := effectstest.B18Counters(t, g, quest, "quest"); n != 1 {
		t.Errorf("quest counters %d, want 1", n)
	}
}

func wwk487NihilUpkeep(t *testing.T, counters, handSize int) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	quest := effectstest.B12Push(g, me.ID, "Quest for the Nihil Stone", wwk487EnchantmentTypeLn, wwk487QuestNihilStone, 0, 0)
	effectstest.AddCountersForTest(g, quest, "quest", counters)
	opp.Hand.Cards = nil
	for i := 0; i < handSize; i++ {
		effectstest.HandCardForTest(opp, "Card", "Sorcery", "")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	return g, opp
}

func TestWwkB487QuestForTheNihilStoneDrainsAnOpponentWithAnEmptyHand(t *testing.T) {
	g, opp := wwk487NihilUpkeep(t, 2, 0)
	me := g.Seats[0]
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-5 {
		t.Errorf("life %d, want %d", got, life-5)
	}
}

func TestWwkB487QuestForTheNihilStoneNeedsTwoCountersAndAnEmptyHand(t *testing.T) {
	for _, tc := range []struct {
		name              string
		counters, handLen int
	}{{"one counter", 1, 0}, {"a card in hand", 2, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			g, opp := wwk487NihilUpkeep(t, tc.counters, tc.handLen)
			me := g.Seats[0]
			life := effectstest.LifeOf(g, opp.ID)
			if effectstest.B19HasTriggerPromptFor(g, me.ID) {
				t.Fatal("the upkeep trigger asked its question")
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.LifeOf(g, opp.ID); got != life {
				t.Errorf("life %d, want %d", got, life)
			}
		})
	}
}

func TestWwkB487QuestForTheNihilStoneMayDecline(t *testing.T) {
	g, opp := wwk487NihilUpkeep(t, 3, 0)
	me := g.Seats[0]
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life {
		t.Errorf("life %d after declining, want %d", got, life)
	}
}

// --- Talus Paladin -----------------------------------------------------------

func TestWwkB487TalusPaladinGivesAlliesLifelinkAndTakesACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ally := effectstest.B12Creature(g, me.ID, "Other Ally", wwk487AllyType, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", wwk487BearType, 2, 2)
	paladin := effectstest.CastHoldCreature(t, g, "Talus Paladin", "Creature — Human Knight Ally", wwk487TalusPaladin, 2, 3)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{ally, paladin} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "lifelink") {
			t.Errorf("an Ally has no lifelink")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("a non-Ally gained lifelink")
	}
	if p, tough := effectstest.PtOf(t, g, paladin); p != 3 || tough != 4 {
		t.Errorf("the Paladin is %d/%d, want 3/4", p, tough)
	}
}

func TestWwkB487TalusPaladinAsksTheCounterEvenWhenLifelinkIsDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	paladin := effectstest.CastHoldCreature(t, g, "Talus Paladin", "Creature — Human Knight Ally", wwk487TalusPaladin, 2, 3)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, paladin, "lifelink") {
		t.Error("lifelink was declined")
	}
	if p, _ := effectstest.PtOf(t, g, paladin); p != 3 {
		t.Errorf("power %d, want 3 (the counter was taken)", p)
	}
}

func TestWwkB487TalusPaladinTriggersForAnotherAlly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	paladin := effectstest.B12Push(g, me.ID, "Talus Paladin", "Creature — Human Knight Ally", wwk487TalusPaladin, 2, 3)
	effectstest.CastHoldCreature(t, g, "New Ally", wwk487AllyType, "", 1, 1)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, paladin); p != 3 {
		t.Errorf("the Paladin has power %d, want 3 (a counter for another Ally entering)", p)
	}
}

// --- Feral Contest -----------------------------------------------------------

func TestWwkB487FeralContestPutsACounterAndForcesABlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", wwk487BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", wwk487BearType, 2, 2)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1}}
	effectstest.CastFor(t, g, me, "Feral Contest", "Sorcery", wwk487FeralContest, game.CastSpellParams{Targets: targets, Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B18Counters(t, g, mine, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters %d, want 1", n)
	}
	if p, tough := effectstest.PtOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("the creature is %d/%d, want 3/3", p, tough)
	}
	if recs := effectstest.BlocksAttackerRecords(g); len(recs) != 1 || recs[0].Objects[0].ID != mine {
		t.Errorf("records = %+v, want one naming the creature that got the counter", recs)
	}
}

func TestWwkB487FeralContestRefusesTheSameCreatureTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "Mine", wwk487BearType, 2, 2)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: mine, Slot: 1}}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Feral Contest", "Sorcery", wwk487FeralContest, targets); err == nil {
		t.Fatal("one creature in both clauses was accepted")
	}
}

func TestWwkB487FeralContestRefusesACreatureYouDoNotControlForTheCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", wwk487BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", wwk487BearType, 2, 2)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: theirs, Slot: 0}, {Kind: game.TargetCard, ID: mine, Slot: 1}}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Feral Contest", "Sorcery", wwk487FeralContest, targets); err == nil {
		t.Fatal("a counter on an opponent's creature was accepted")
	}
}
