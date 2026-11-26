package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b493_cards_test.go — Zendikar quest enchantments, S82 OLD batch 493
// (#5583), each driven through the engine.

const (
	zenHolyRelic   = "0fa5803b-3604-46c7-9774-bf70e4a2b925"
	zenTrickster   = "372633d9-ac74-4630-a1ae-a1906bb2aed0"
	zenArchmage    = "5966c464-5f63-4c71-9317-e2ca79da5aba"
	zenPyromancer  = "85845b6c-abb5-4591-b987-a375972cf25b"
	zenGemblades   = "c762361a-0b72-4db5-8b6f-715ca42d5067"
	zenAncientSecr = "d7f7b882-90cb-4031-8a57-8e721bc905ac"
)

func zenAddCounters(g *game.Game, id uuid.UUID, kind string, n int) {
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, kind, n) })
}

func TestS82OldB493EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		zenHolyRelic: "Quest for the Holy Relic", zenTrickster: "Cosi's Trickster", zenArchmage: "Archmage Ascension",
		zenPyromancer: "Pyromancer Ascension", zenGemblades: "Quest for the Gemblades", zenAncientSecr: "Quest for Ancient Secrets",
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

func TestS82OldB493HolyRelicCountsCreatureSpellsThenFetchesAndAttachesEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	quest := effectstest.PushCatalogPermanent(g, me.ID, "Quest for the Holy Relic", "Enchantment", zenHolyRelic, false)
	effectstest.FillPool(me, 1)
	effectstest.CastFromHandForTest(t, g, me, "Bear Cub", "Creature — Bear", "{1}", "", game.CastSpellParams{Strict: true})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, quest, "quest"); got != 1 {
		t.Fatalf("%d quest counters after a creature spell, want 1", got)
	}
	// A noncreature spell adds nothing.
	effectstest.FillPool(me, 1)
	effectstest.CastFromHandForTest(t, g, me, "Trinket", "Artifact", "{1}", "", game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, quest, "quest"); got != 1 {
		t.Fatalf("a noncreature spell changed the counters to %d", got)
	}

	// Four short: the activation cannot be paid.
	zenAddCounters(g, quest, "quest", 3)
	if err := g.ActivateCatalogAbility(me.ID, quest, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with four quest counters")
	}
	zenAddCounters(g, quest, "quest", 1)

	bear := effectstest.SeedBear(g, me.ID)
	other := effectstest.SeedBear(g, me.ID)
	blade := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: blade, Name: "Test Blade", TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID})
	effectstest.ApaActivate(t, g, me, quest, 0, game.ActivateAbilityParams{})
	effectstest.AnswerSearchByID(t, g, me.ID, blade)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(quest) {
		t.Error("the quest was not sacrificed")
	}
	if !g.Battlefield.Contains(blade) {
		t.Fatal("the Equipment did not reach the battlefield")
	}
	if !effectstest.AttachedTo(t, g, blade, other) {
		t.Error("the Equipment is not attached to the chosen creature")
	}
	if effectstest.AttachedTo(t, g, blade, bear) {
		t.Error("the Equipment is attached to the other creature too")
	}
}

func TestS82OldB493CosisTricksterGrowsOnAnOpponentsShuffleOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	trickster := effectstest.PushCatalogPermanent(g, me.ID, "Cosi's Trickster", "Creature — Merfolk Wizard", zenTrickster, false)
	g.WithWriteLock(func() { _ = g.ShuffleLibraryForEffect(me.ID) })
	if effectstest.LatestTriggerPromptFor(g, me.ID) != nil {
		t.Fatal("the Trickster triggered on its controller's own shuffle")
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.ShuffleLibraryForEffect(opp.ID) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, trickster, game.CounterPlusOne); got != 0 {
		t.Fatalf("declined, the Trickster has %d counters", got)
	}
	g.WithWriteLock(func() { _ = g.ShuffleLibraryForEffect(opp.ID) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, trickster, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d +1/+1 counters, want 1", got)
	}
	if p, tough := effectstest.PtOf(t, g, trickster); p != 2 || tough != 2 {
		t.Errorf("the Trickster is %d/%d, want 2/2", p, tough)
	}
}

func TestS82OldB493ArchmageAscensionCountsTwoDrawsAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	quest := effectstest.PushCatalogPermanent(g, me.ID, "Archmage Ascension", "Enchantment", zenArchmage, false)
	g.WithWriteLock(func() { g.DrawnThisTurn[me.ID] = nil })
	effectstest.DrawN(g, me, 1)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPromptFor(g, me.ID) != nil {
		t.Fatal("the quest triggered after one draw")
	}
	if got := effectstest.CountersOn(g, quest, "quest"); got != 0 {
		t.Fatalf("%d counters after one draw", got)
	}
}

func TestS82OldB493ArchmageAscensionPutsACounterAfterTwoDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	quest := effectstest.PushCatalogPermanent(g, me.ID, "Archmage Ascension", "Enchantment", zenArchmage, false)
	g.WithWriteLock(func() { g.DrawnThisTurn[me.ID] = nil })
	effectstest.DrawN(g, me, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, quest, "quest"); got != 1 {
		t.Fatalf("%d counters after two draws, want 1", got)
	}
}

func TestS82OldB493ArchmageAscensionSearchesInsteadOfDrawingAtSixCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	quest := effectstest.PushCatalogPermanent(g, me.ID, "Archmage Ascension", "Enchantment", zenArchmage, false)
	zenAddCounters(g, quest, "quest", 5)
	// Five counters: an ordinary draw, no question.
	hand := me.Hand.Size()
	effectstest.DrawN(g, me, 1)
	if effectstest.OptionalOffer(g, me) != nil {
		t.Fatal("asked to replace a draw with five counters")
	}
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
	zenAddCounters(g, quest, "quest", 1)
	wanted := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: wanted, Name: "Wanted", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	hand = me.Hand.Size()
	// Declined: the top card is drawn.
	effectstest.DrawN(g, me, 1)
	effectstest.AnswerOptional(t, g, me, false)
	if me.Hand.Size() != hand+1 || me.Hand.Contains(wanted) {
		t.Fatal("declining did not draw the top card")
	}
	// Accepted: search instead.
	hand = me.Hand.Size()
	effectstest.DrawN(g, me, 1)
	effectstest.AnswerOptional(t, g, me, true)
	effectstest.AnswerSearchByID(t, g, me.ID, wanted)
	if !me.Hand.Contains(wanted) || me.Hand.Size() != hand+1 {
		t.Fatalf("the searched card is not the only addition to hand (hand %d -> %d)", hand, me.Hand.Size())
	}
}

func TestS82OldB493PyromancerAscensionCountsSameNameSpellsThenCopies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	quest := effectstest.PushCatalogPermanent(g, me.ID, "Pyromancer Ascension", "Enchantment", zenPyromancer, false)
	// No same-named card in the graveyard: no trigger.
	effectstest.FillPool(me, 1)
	effectstest.CastFromHandForTest(t, g, me, "Spark", "Instant", "{1}", "", game.CastSpellParams{Strict: true})
	if effectstest.LatestTriggerPromptFor(g, me.ID) != nil {
		t.Fatal("triggered with no same-named card in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Spark is in the graveyard now: another Spark counts.
	for i := 0; i < 2; i++ {
		effectstest.FillPool(me, 1)
		effectstest.CastFromHandForTest(t, g, me, "Spark", "Instant", "{1}", "", game.CastSpellParams{Strict: true})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CountersOn(g, quest, "quest"); got != 2 {
		t.Fatalf("%d quest counters, want 2", got)
	}
	// Two counters: any instant is copied, on a yes.
	effectstest.FillPool(me, 1)
	effectstest.CastFromHandForTest(t, g, me, "Other Spell", "Sorcery", "{1}", "", game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	if got := effectstest.AcCopiesOnStack(g); got != 1 {
		t.Fatalf("%d copies on the stack, want 1", got)
	}
}

func TestS82OldB493GembladesCountsCombatDamageToCreaturesThenPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	quest := effectstest.PushCatalogPermanent(g, me.ID, "Quest for the Gemblades", "Enchantment", zenGemblades, false)
	// Noncombat damage to a creature is not counted.
	pinger := effectstest.B12Creature(g, me.ID, "Pinger", "Creature — Bear", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	effectstest.P2Ping(t, g, pinger, victim, 1)
	if effectstest.LatestTriggerPromptFor(g, me.ID) != nil {
		t.Fatal("triggered on noncombat damage")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Combat damage to a creature is.
	effectstest.Pr7bAttack(t, g, opp.ID, pinger)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{victim: pinger})
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, quest, "quest"); got != 1 {
		t.Fatalf("%d quest counters, want 1", got)
	}
	target := effectstest.B12Creature(g, me.ID, "Target", "Creature — Bear", 1, 1)
	effectstest.ApaActivate(t, g, me, quest, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(target)})
	if g.Battlefield.Contains(quest) {
		t.Error("the quest was not sacrificed")
	}
	if got := effectstest.CountersOn(g, target, game.CounterPlusOne); got != 4 {
		t.Errorf("%d +1/+1 counters, want 4", got)
	}
}

func TestS82OldB493AncientSecretsCountsEveryRouteIntoYourGraveyardThenShuffles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	quest := effectstest.PushCatalogPermanent(g, me.ID, "Quest for Ancient Secrets", "Enchantment", zenAncientSecr, false)
	// An opponent's card does not count.
	theirs := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: theirs, Name: "Theirs", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() { _ = g.DiscardCardsKeyedForEffect(opp.ID, []uuid.UUID{theirs}, uuid.Nil, game.KeyedTail{}) })
	if effectstest.LatestTriggerPromptFor(g, me.ID) != nil {
		t.Fatal("triggered on an opponent's graveyard")
	}
	// A discard, a mill and a dying creature each count.
	mine := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: mine, Name: "Mine", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { _ = g.DiscardCardsKeyedForEffect(me.ID, []uuid.UUID{mine}, uuid.Nil, game.KeyedTail{}) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, quest, "quest"); got != 2 {
		t.Fatalf("%d quest counters, want 2", got)
	}
	zenAddCounters(g, quest, "quest", 3)
	libBefore := len(opp.Library.Cards)
	gyBefore := len(opp.Graveyard.Cards)
	if gyBefore == 0 {
		t.Fatal("test setup: the opponent's graveyard is empty")
	}
	effectstest.ApaActivate(t, g, me, quest, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if g.Battlefield.Contains(quest) {
		t.Error("the quest was not sacrificed")
	}
	if len(opp.Graveyard.Cards) != 0 || len(opp.Library.Cards) != libBefore+gyBefore {
		t.Errorf("graveyard %d, library %d -> %d: the graveyard was not shuffled in", len(opp.Graveyard.Cards), libBefore, len(opp.Library.Cards))
	}
}
