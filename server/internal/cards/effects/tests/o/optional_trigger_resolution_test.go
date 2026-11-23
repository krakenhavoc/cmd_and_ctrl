package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The issue's report: Kodama's Reach resolving with Avenger of Zendikar
// out. The first search puts a land onto the battlefield, and the
// Avenger's landfall used to ask its "you may" right then, beside the
// spell's second search. Now only the spell's own prompts are open
// until it finishes; the landfall trigger goes on the stack after it
// and asks as it resolves.
func TestKodamasReachAsksOnlyItsOwnSearchesWithAvengerOut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	var forests []uuid.UUID
	for i := 0; i < 3; i++ {
		forests = append(forests, effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}))
	}
	avenger := effectstest.PushCatalogPermanent(g, me.ID, "Avenger of Zendikar", "Creature — Elemental", effectstest.AvengerOfZendikarOracle, false)
	plant := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Plant", TypeLine: "Token Creature — Plant", Power: 0, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})

	effectstest.CastCatalogSpell(t, g, "Kodama's Reach", "Sorcery — Arcane", effectstest.C2875KodamaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceTriggerPrompt) != nil {
		t.Fatal("a \"you may\" is open before the spell's first search")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, forests[0])
	if !g.Battlefield.Contains(forests[0]) {
		t.Fatal("the first search did not put the Forest onto the battlefield")
	}
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceTriggerPrompt) != nil {
		t.Fatal("the Avenger asked its \"you may\" while Kodama's Reach was still resolving")
	}
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("the second search is not the open prompt")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, forests[1])

	// The spell has finished: the landfall trigger is on the stack,
	// carrying its question, and nothing has been asked yet.
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceTriggerPrompt) != nil {
		t.Fatal("the Avenger asked before its trigger went on the stack")
	}
	g.RunStateChecksForTest()
	item := effectstest.TriggerOnStack(g, avenger)
	if item == nil || item.Optional == nil {
		t.Fatalf("the landfall trigger is not on the stack with its \"you may\": %+v", item)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, plant, "+1/+1"); got != 1 {
		t.Errorf("the Plant has %d +1/+1 counters, want 1", got)
	}
}

// A response between the trigger going on the stack and its "you may":
// Eternal Witness's chosen card leaves the graveyard, so the ability is
// countered on resolution (CR 608.2b) and nothing is asked.
func TestEternalWitnessTargetGoneInResponseAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[0]
	gone := effectstest.PushGraveyardCardForTest(caster, "Gone Body")
	effectstest.PushGraveyardCardForTest(caster, "Other Body")

	witnessID := effectstest.CastAndResolveCreature(t, g, "Eternal Witness", "Creature — Human Shaman", effectstest.EternalWitnessOracle)
	effectstest.PickCard(t, g, caster.ID, gone)
	if effectstest.TriggerOnStack(g, witnessID) == nil {
		t.Fatal("the trigger is not on the stack after its target pick")
	}
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(gone); err != nil {
			t.Fatalf("exile in response: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, caster.ID) != nil {
		t.Error("a trigger countered on resolution still asked its \"you may\"")
	}
	if caster.Hand.Contains(gone) {
		t.Error("the exiled card came back")
	}
}

// Reclamation Sage, declined as it resolves: the chosen artifact stays.
func TestReclamationSageDeclinedAtResolutionDestroysNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	sword := effectstest.PushTypedCard(g, opp.ID, "Sword", "Artifact", "{1}")
	effectstest.PushTypedCard(g, opp.ID, "Other", "Artifact", "{1}")
	sageID := effectstest.CastAndResolveCreature(t, g, "Reclamation Sage", "Creature — Elf Shaman", effectstest.ReclamationSageOracle)
	effectstest.PickCard(t, g, caster.ID, sword)
	if effectstest.TriggerOnStack(g, sageID) == nil {
		t.Fatal("the trigger is not on the stack after its target pick")
	}
	ask := effectstest.AwaitTriggerPrompt(t, g, caster.ID)
	if effectstest.TriggerOnStack(g, sageID) != nil {
		t.Error("the trigger is still waiting on the stack while it asks")
	}
	if err := g.ResolveTriggerPrompt(ask.ID, caster.ID, false); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sword) {
		t.Error("the declined trigger destroyed its target")
	}
	if len(g.PendingChoices) != 0 || !effectstest.StackFullyEmpty(g) {
		t.Error("the declined trigger did not finish resolving")
	}
}

// An optional trigger its controller has a standing "No" for resolves
// doing nothing, so it is not ordered against a footprinted batch-mate
// (ADR 0197 §7). Without the standing answer the same batch asks.
func TestStandingNoOptionalTriggerIsNotOrdered(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		effectstest.PushCatalogPermanent(g, me.ID, "Independence Token Probe", "Enchantment", effectstest.IndependenceTokenProbe, false)
		effectstest.PushCatalogPermanent(g, me.ID, "Optional Draw Probe", "Enchantment", effectstest.OptionalDrawOnCastProbe, false)
		return g, me
	}
	optionalKey := func(g *game.Game) string {
		for _, it := range g.PendingTriggers {
			if it != nil && it.Optional != nil {
				return it.Optional.AnswerKey
			}
		}
		return ""
	}

	g, _ := setup(t)
	effectstest.CastNoncreature(t, g)
	if effectstest.SettleUntilOrderPrompt(t, g) == nil {
		t.Fatal("setup: an optional trigger and a token trigger are not offered for ordering")
	}
	key := optionalKey(g)
	if key == "" {
		t.Fatal("the optional trigger carries no standing-answer key")
	}

	g, me := setup(t)
	if err := g.SetAutoAnswers(me.ID, map[string]game.AutoAnswer{key: game.AutoAnswerNever}); err != nil {
		t.Fatal(err)
	}
	hand := me.Hand.Size()
	effectstest.CastNoncreature(t, g)
	if ch := effectstest.SettleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("a batch whose optional trigger will be declined asked to be ordered: %+v", ch)
	}
	ask := effectstest.AwaitTriggerPrompt(t, g, me.ID)
	if ask.AutoAnswerKey != key {
		t.Fatalf("prompt key %q, want %q", ask.AutoAnswerKey, key)
	}
	if n := effectstest.AutoAnswerAll(t, g); n != 1 {
		t.Fatalf("answered %d automatically, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("hand moved by %d, want 0: declined, nothing drawn", got)
	}
}

// landTaxTable seats Land Tax for seat 1 with an opponent ahead on
// lands, at seat 1's upkeep with the trigger on the stack, and a
// standing "Always" for it (ADR 0127).
func landTaxTable(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Land Tax", "Enchantment", effectstest.B03LandTaxOracle, false)
	effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	for i := 0; i < 3; i++ {
		effectstest.StapleLibraryCard(me, "Plains", "Basic Land — Plains")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	g.RunStateChecksForTest()
	var item *game.StackItem
	for _, it := range g.StackMeta {
		if it != nil && it.Optional != nil {
			item = it
		}
	}
	if item == nil {
		t.Fatal("setup: Land Tax's trigger is not on the stack")
	}
	if err := g.SetAutoAnswers(me.ID, map[string]game.AutoAnswer{item.Optional.AnswerKey: game.AutoAnswerAlways}); err != nil {
		t.Fatal(err)
	}
	return g, me
}

// #3907, CR 603.4 before CR 603.5: the condition fails in response, so
// the trigger is removed as it resolves with no question — nothing for
// the standing answer to answer — and the log says why.
func TestOptionalTriggerWhoseConditionFailsAsksNothing(t *testing.T) {
	g, me := landTaxTable(t)
	effectstest.SeedLandOnBattlefield(g, me.ID, "Plains", "Basic Land — Plains") // evens the count
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a trigger whose condition failed asked its \"you may\"")
	}
	if n := effectstest.AutoAnswerAll(t, g); n != 0 {
		t.Errorf("answered %d automatically, want 0: nothing was asked", n)
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil || me.Hand.Size() != hand {
		t.Error("the trigger did something")
	}
	var why *game.Event
	for i := range g.Events {
		switch g.Events[i].Kind {
		case game.EventFizzle:
			why = &g.Events[i]
		case game.EventAutoAnswer:
			t.Error("an automatic answer was recorded for a question never asked")
		}
	}
	if why == nil || why.Call != game.FizzleCallCondition {
		t.Fatalf("no fizzle marked as a failed condition: %+v", why)
	}
	if !effectstest.StackFullyEmpty(g) || len(g.PendingChoices) != 0 {
		t.Error("the trigger did not leave the stack")
	}
}

// The condition still holds: the question is asked as before, and the
// standing answer answers it.
func TestOptionalTriggerWhoseConditionHoldsStillAsks(t *testing.T) {
	g, me := landTaxTable(t)
	ask := effectstest.AwaitTriggerPrompt(t, g, me.ID)
	if ask.AutoAnswerKey == "" {
		t.Fatal("the prompt carries no standing-answer key")
	}
	if n := effectstest.AutoAnswerAll(t, g); n != 1 {
		t.Fatalf("answered %d automatically, want 1", n)
	}
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Error("Always did not run the search")
	}
	for i := range g.Events {
		if g.Events[i].Kind == game.EventFizzle {
			t.Error("a trigger whose condition held was removed")
		}
	}
}
