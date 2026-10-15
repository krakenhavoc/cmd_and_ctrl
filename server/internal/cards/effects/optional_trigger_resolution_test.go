package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// optional_trigger_resolution_test.go — #2494, ADR 0197: an optional
// triggered ability goes on the stack like any other and asks its "you
// may" as it resolves (CR 603.5).

const (
	avengerOfZendikarOracle    = "4ba5b3f6-503b-43e6-b66e-4f8c55cffed7"
	optionalDrawOnCastProbe    = "test-2494-optional-draw-on-cast"
	optionalDrawOnCastLabel    = "Optional Draw Probe — you may draw a card"
	optionalDrawOnCastQuestion = "Optional Draw Probe — draw a card?"
)

func init() {
	// "Whenever you cast a noncreature spell, you may draw a card": an
	// optional trigger beside the #2884 token probe's footprinted one.
	Register(Spec{
		OracleID: optionalDrawOnCastProbe,
		Name:     "Optional Draw Probe",
		Triggered: []game.TriggeredAbility{
			Optional(WheneverYouCast(Noncreature(), optionalDrawOnCastLabel, Do(DrawCards{N: 1})),
				optionalDrawOnCastQuestion),
		},
	})
}

// The issue's report: Kodama's Reach resolving with Avenger of Zendikar
// out. The first search puts a land onto the battlefield, and the
// Avenger's landfall used to ask its "you may" right then, beside the
// spell's second search. Now only the spell's own prompts are open
// until it finishes; the landfall trigger goes on the stack after it
// and asks as it resolves.
func TestKodamasReachAsksOnlyItsOwnSearchesWithAvengerOut(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	var forests []uuid.UUID
	for i := 0; i < 3; i++ {
		forests = append(forests, pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}))
	}
	avenger := pushCatalogPermanent(g, me.ID, "Avenger of Zendikar", "Creature — Elemental", avengerOfZendikarOracle, false)
	plant := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Plant", TypeLine: "Token Creature — Plant", Power: 0, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})

	castCatalogSpell(t, g, "Kodama's Reach", "Sorcery — Arcane", c2875KodamaOracle, nil)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKind(g, game.PendingChoiceTriggerPrompt) != nil {
		t.Fatal("a \"you may\" is open before the spell's first search")
	}
	answerSearchByID(t, g, me.ID, forests[0])
	if !g.Battlefield.Contains(forests[0]) {
		t.Fatal("the first search did not put the Forest onto the battlefield")
	}
	if latestChoiceOfKind(g, game.PendingChoiceTriggerPrompt) != nil {
		t.Fatal("the Avenger asked its \"you may\" while Kodama's Reach was still resolving")
	}
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("the second search is not the open prompt")
	}
	answerSearchByID(t, g, me.ID, forests[1])

	// The spell has finished: the landfall trigger is on the stack,
	// carrying its question, and nothing has been asked yet.
	if latestChoiceOfKind(g, game.PendingChoiceTriggerPrompt) != nil {
		t.Fatal("the Avenger asked before its trigger went on the stack")
	}
	g.RunStateChecksForTest()
	item := triggerOnStack(g, avenger)
	if item == nil || item.Optional == nil {
		t.Fatalf("the landfall trigger is not on the stack with its \"you may\": %+v", item)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, plant, "+1/+1"); got != 1 {
		t.Errorf("the Plant has %d +1/+1 counters, want 1", got)
	}
}

// A response between the trigger going on the stack and its "you may":
// Eternal Witness's chosen card leaves the graveyard, so the ability is
// countered on resolution (CR 608.2b) and nothing is asked.
func TestEternalWitnessTargetGoneInResponseAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	gone := pushGraveyardCardForTest(caster, "Gone Body")
	pushGraveyardCardForTest(caster, "Other Body")

	witnessID := castAndResolveCreature(t, g, "Eternal Witness", "Creature — Human Shaman", eternalWitnessOracle)
	pickCard(t, g, caster.ID, gone)
	if triggerOnStack(g, witnessID) == nil {
		t.Fatal("the trigger is not on the stack after its target pick")
	}
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(gone); err != nil {
			t.Fatalf("exile in response: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, caster.ID) != nil {
		t.Error("a trigger countered on resolution still asked its \"you may\"")
	}
	if caster.Hand.Contains(gone) {
		t.Error("the exiled card came back")
	}
}

// Reclamation Sage, declined as it resolves: the chosen artifact stays.
func TestReclamationSageDeclinedAtResolutionDestroysNothing(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	sword := pushTypedCard(g, opp.ID, "Sword", "Artifact", "{1}")
	pushTypedCard(g, opp.ID, "Other", "Artifact", "{1}")
	sageID := castAndResolveCreature(t, g, "Reclamation Sage", "Creature — Elf Shaman", reclamationSageOracle)
	pickCard(t, g, caster.ID, sword)
	if triggerOnStack(g, sageID) == nil {
		t.Fatal("the trigger is not on the stack after its target pick")
	}
	ask := awaitTriggerPrompt(t, g, caster.ID)
	if triggerOnStack(g, sageID) != nil {
		t.Error("the trigger is still waiting on the stack while it asks")
	}
	if err := g.ResolveTriggerPrompt(ask.ID, caster.ID, false); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sword) {
		t.Error("the declined trigger destroyed its target")
	}
	if len(g.PendingChoices) != 0 || !stackFullyEmpty(g) {
		t.Error("the declined trigger did not finish resolving")
	}
}

// An optional trigger its controller has a standing "No" for resolves
// doing nothing, so it is not ordered against a footprinted batch-mate
// (ADR 0197 §7). Without the standing answer the same batch asks.
func TestStandingNoOptionalTriggerIsNotOrdered(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		pushCatalogPermanent(g, me.ID, "Independence Token Probe", "Enchantment", independenceTokenProbe, false)
		pushCatalogPermanent(g, me.ID, "Optional Draw Probe", "Enchantment", optionalDrawOnCastProbe, false)
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
	castNoncreature(t, g)
	if settleUntilOrderPrompt(t, g) == nil {
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
	castNoncreature(t, g)
	if ch := settleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("a batch whose optional trigger will be declined asked to be ordered: %+v", ch)
	}
	ask := awaitTriggerPrompt(t, g, me.ID)
	if ask.AutoAnswerKey != key {
		t.Fatalf("prompt key %q, want %q", ask.AutoAnswerKey, key)
	}
	if n := autoAnswerAll(t, g); n != 1 {
		t.Fatalf("answered %d automatically, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("hand moved by %d, want 0: declined, nothing drawn", got)
	}
}
