package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// retriever_phoenix_test.go — Retriever Phoenix (#3525): learn is an
// event a replacement effect can apply to (CR 701.48a, CR 614.1a,
// game/learn_action.go), and the Phoenix's graveyard clause is one.
// Learn is driven through Cram Session (learn_test.go's table), whose
// only other clause is a life gain.

const (
	retrieverPhoenixOracle = "279e573d-b7a3-4e62-a708-3ec992d97972"
	phoenixTypeLine        = "Creature — Phoenix"
)

func gyPhoenix(p *game.Player) uuid.UUID {
	return effectstest.GyCardWithOracle(p, "Retriever Phoenix", phoenixTypeLine, retrieverPhoenixOracle)
}

func phoenixInPlay(g *game.Game, id uuid.UUID) bool { return g.Battlefield.Contains(id) }

// learnOffer is the Phoenix's yes/no.
func learnOffer(g *game.Game, p *game.Player) *game.PendingChoice {
	return effectstest.OptionalOffer(g, p)
}

// A learn with a Phoenix in the learner's graveyard asks first, before
// any of the learn itself: nothing has been discarded or looked at.
func TestRetrieverPhoenixIsOfferedBeforeTheLearnHappens(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.LessonCard("Environmental Sciences"))
	me := g.Seats[0]
	phoenix := gyPhoenix(me)
	effectstest.CastCatalogSpell(t, g, "Cram Session", "Sorcery", effectstest.CramSessionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	offer := learnOffer(g, me)
	if offer == nil {
		t.Fatalf("the Phoenix was not offered: %+v", g.PendingChoices)
	}
	if offer.Source != phoenix {
		t.Errorf("the offer names %v, want the Phoenix %v", offer.Source, phoenix)
	}
	if effectstest.LearnDiscardPrompt(g, me) != nil || effectstest.LearnLessonPrompt(g, me) != nil {
		t.Error("the learn's own questions were asked before the replacement was answered")
	}
}

// Yes: the learn never happens (CR 614.6) and the Phoenix returns. The
// discard, the draw and the Lesson are not looked at, and the Phoenix is
// not "cast", so its own enters trigger does not learn.
func TestRetrieverPhoenixReturnsInsteadOfLearning(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.LessonCard("Environmental Sciences"))
	me := g.Seats[0]
	phoenix := gyPhoenix(me)
	handBefore := me.Hand.Size()
	lifeBefore := me.Life
	effectstest.CastCatalogSpell(t, g, "Cram Session", "Sorcery", effectstest.CramSessionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AnswerOptional(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)

	if !phoenixInPlay(g, phoenix) || effectstest.ZoneHas(me.Graveyard, phoenix) {
		t.Fatal("the Phoenix did not return to the battlefield")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("a replaced learn left a question open: %+v", g.PendingChoices)
	}
	// The spell cast to learn left the hand; nothing was discarded or drawn.
	if me.Hand.Size() != handBefore {
		t.Errorf("hand %d -> %d: a replaced learn discards and draws nothing (Cram Session left it)", handBefore, me.Hand.Size())
	}
	if me.Life != lifeBefore+4 {
		t.Errorf("life %d -> %d: only the learn is replaced, Cram Session still gains 4", lifeBefore, me.Life)
	}
	if got := effectstest.SideNamed(me); !effectstest.EqualNames(got, []string{"Environmental Sciences"}) {
		t.Errorf("sideboard %v: a replaced learn takes no Lesson", got)
	}
	card, _ := g.LookupCardForEffect(phoenix)
	if card.Controller != me.ID {
		t.Errorf("the Phoenix came back under %v, want its owner %v", card.Controller, me.ID)
	}
}

// No: "you may instead". The learn goes ahead as printed, the Phoenix
// stays where it is, and the offer is asked once.
func TestDecliningRetrieverPhoenixLearnsAsPrinted(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.LessonCard("Environmental Sciences"))
	me := g.Seats[0]
	phoenix := gyPhoenix(me)
	effectstest.CastCatalogSpell(t, g, "Cram Session", "Sorcery", effectstest.CramSessionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AnswerOptional(t, g, me, false)
	if learnOffer(g, me) != nil {
		t.Error("the offer was asked twice for one learn (CR 614.5)")
	}
	if effectstest.LearnDiscardPrompt(g, me) == nil {
		t.Fatalf("a declined replacement did not go on to the learn: %+v", g.PendingChoices)
	}
	if !effectstest.ZoneHas(me.Graveyard, phoenix) || phoenixInPlay(g, phoenix) {
		t.Error("a declined Phoenix must stay in the graveyard")
	}
	// And the learn is whole: a player who discards draws.
	c := effectstest.LearnDiscardPrompt(g, me)
	pitched := c.ChooseCards[0]
	handBefore := me.Hand.Size()
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{pitched}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ZoneHas(me.Graveyard, pitched) || me.Hand.Size() != handBefore {
		t.Errorf("discard then draw: graveyard has it %v, hand %d -> %d", effectstest.ZoneHas(me.Graveyard, pitched), handBefore, me.Hand.Size())
	}
}

// Two Phoenixes in one graveyard are two "may"s asked in turn, so the
// player takes one or none, never both (CR 616.1: the player chooses
// which to apply, and the learn is gone after the first).
func TestTwoRetrieverPhoenixesReturnOneOrNone(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	a, b := gyPhoenix(me), gyPhoenix(me)
	effectstest.CastCatalogSpell(t, g, "Cram Session", "Sorcery", effectstest.CramSessionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	first := learnOffer(g, me)
	if first == nil {
		t.Fatal("no offer")
	}
	effectstest.AnswerOptional(t, g, me, false)
	second := learnOffer(g, me)
	if second == nil {
		t.Fatal("the second Phoenix was never offered")
	}
	if second.Source == first.Source {
		t.Fatal("the same Phoenix was offered twice")
	}
	effectstest.AnswerOptional(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)

	took, left := second.Source, first.Source
	if !phoenixInPlay(g, took) || !effectstest.ZoneHas(me.Graveyard, left) {
		t.Errorf("only the accepted Phoenix returns: took %v on battlefield %v, left %v in graveyard %v",
			took == a || took == b, phoenixInPlay(g, took), left == a || left == b, effectstest.ZoneHas(me.Graveyard, left))
	}
	if learnOffer(g, me) != nil || effectstest.LearnDiscardPrompt(g, me) != nil {
		t.Error("a learn is replaced by at most one Phoenix, and then it is over")
	}
}

// Both can say no, and then the learn happens.
func TestTwoRetrieverPhoenixesBothDeclinedLearnAsPrinted(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	a, b := gyPhoenix(me), gyPhoenix(me)
	effectstest.CastCatalogSpell(t, g, "Cram Session", "Sorcery", effectstest.CramSessionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptional(t, g, me, false)
	effectstest.AnswerOptional(t, g, me, false)
	if effectstest.LearnDiscardPrompt(g, me) == nil {
		t.Fatalf("no learn after declining both: %+v", g.PendingChoices)
	}
	if !effectstest.ZoneHas(me.Graveyard, a) || !effectstest.ZoneHas(me.Graveyard, b) {
		t.Error("a declined Phoenix leaves the graveyard")
	}
}

// "As long as this card is in YOUR graveyard, if YOU would learn": an
// opponent's graveyard Phoenix does not apply to my learn, and a
// Phoenix anywhere but a graveyard does not apply at all.
func TestRetrieverPhoenixAppliesOnlyFromTheLearnersGraveyard(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me, opp := g.Seats[0], g.Seats[1]
	gyPhoenix(opp)
	effectstest.BfCardWithOracle(g, me, "Retriever Phoenix", phoenixTypeLine, retrieverPhoenixOracle)
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Retriever Phoenix", TypeLine: phoenixTypeLine,
		OracleID: retrieverPhoenixOracle, Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Cram Session", "Sorcery", effectstest.CramSessionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if learnOffer(g, me) != nil || learnOffer(g, opp) != nil {
		t.Fatalf("a Phoenix off the learner's graveyard was offered: %+v", g.PendingChoices)
	}
	if effectstest.LearnDiscardPrompt(g, me) == nil {
		t.Errorf("the learn did not go ahead: %+v", g.PendingChoices)
	}
}

// Another player's learn is their own: a Phoenix in my graveyard is not
// offered to the opponent who learns (the replacement is "you").
func TestRetrieverPhoenixIsNotOfferedToAnotherPlayersLearn(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me, opp := g.Seats[0], g.Seats[1]
	gyPhoenix(me)
	g.WithWriteLock(func() { _ = g.LearnForEffect(opp.ID, uuid.Nil) })
	if learnOffer(g, me) != nil || learnOffer(g, opp) != nil {
		t.Errorf("a Phoenix was offered on an opponent's learn: %+v", g.PendingChoices)
	}
}

// More than one replacement of a learn: the learner chooses the order
// (CR 616.1). A second effect that is not a graveyard "may" puts the
// ordering prompt up, and the order chosen is the order applied: with
// the Phoenix first and "yes" the learn is gone and the other effect has
// nothing to replace.
func TestRetrieverPhoenixAndAnotherLearnReplacementAreOrdered(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	phoenix := gyPhoenix(me)
	other := game.ReplacementEffect{
		Watches: []game.EventKind{game.EventKeywordAction},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
			return ev.Kind == game.RepEventKeywordAction && ev.KeywordAction == game.KeywordActionLearn
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.Cancel()
			return nil
		},
		Label: "learn is skipped",
	}
	g.RegisterReplacementForTest(other)
	effectstest.CastCatalogSpell(t, g, "Cram Session", "Sorcery", effectstest.CramSessionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	order := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceReplacementOrder, me.ID)
	if order == nil {
		t.Fatalf("two replacements of one learn must ask the learner to order them: %+v", g.PendingChoices)
	}
	var phoenixFirst []game.ReplacementEffectID
	for _, id := range order.ReplacementEffectIDs {
		if _, card := g.ReplacementOptionMetaForEffect(id); card == phoenix {
			phoenixFirst = append([]game.ReplacementEffectID{id}, phoenixFirst...)
		} else {
			phoenixFirst = append(phoenixFirst, id)
		}
	}
	if len(phoenixFirst) != 2 {
		t.Fatalf("%d effects in the order prompt, want 2", len(phoenixFirst))
	}
	if err := g.ResolveReplacementOrder(order.ID, me.ID, phoenixFirst); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	effectstest.AnswerOptional(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !phoenixInPlay(g, phoenix) {
		t.Error("the Phoenix, ordered first and accepted, did not return")
	}
}

// Cast from the hand, the Phoenix learns when it enters: the front half
// is the same learn.
func TestRetrieverPhoenixLearnsWhenCast(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.LessonCard("Environmental Sciences"))
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Retriever Phoenix", phoenixTypeLine, retrieverPhoenixOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LearnDiscardPrompt(g, me) == nil {
		t.Fatalf("a cast Phoenix did not learn: %+v", g.PendingChoices)
	}
}

// "If you cast it": a Phoenix that is returned does not learn. This is
// the guard that keeps the graveyard clause from feeding itself.
func TestReturnedRetrieverPhoenixDoesNotLearn(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.LessonCard("Environmental Sciences"))
	me := g.Seats[0]
	phoenix := gyPhoenix(me)
	g.WithWriteLock(func() { _ = g.ReturnFromGraveyardForEffect(phoenix, game.ZoneBattlefield) })
	effectstest.PassPriorityAroundTable(t, g)
	if !phoenixInPlay(g, phoenix) {
		t.Fatal("the Phoenix did not return")
	}
	if effectstest.LearnDiscardPrompt(g, me) != nil || learnOffer(g, me) != nil || len(g.PendingChoices) != 0 {
		t.Errorf("a returned Phoenix learned: %+v", g.PendingChoices)
	}
}

// A card that left the graveyard while the question was open (nothing
// can act then, but the body still checks) is not returned: it reads the
// zone when the window settles.
func TestRetrieverPhoenixBodyChecksTheGraveyard(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	phoenix := gyPhoenix(me)
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(phoenix, game.ZoneHand); err != nil {
			t.Fatal(err)
		}
		done := false
		if err := effects.ReturnThisInsteadOfLearningBody(g, me.ID, phoenix, func(*game.Game) error { done = true; return nil }); err != nil {
			t.Fatal(err)
		}
		if !done {
			t.Error("the body did not finish the sentence")
		}
	})
	if phoenixInPlay(g, phoenix) || !effectstest.ZoneHas(me.Hand, phoenix) {
		t.Error("a Phoenix no longer in the graveyard was reanimated from the hand")
	}
}

// A table with a Phoenix in the graveyard is a restore point (nothing is
// paused), and the restored game still offers the replacement.
func TestRetrieverPhoenixInTheGraveyardSurvivesARestore(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.LessonCard("Environmental Sciences"))
	me := g.Seats[0]
	gyPhoenix(me)
	restored := effectstest.RestoreThroughJSON(t, g)
	rme := restored.Seats[0]
	effectstest.CastCatalogSpell(t, restored, "Cram Session", "Sorcery", effectstest.CramSessionOracle, nil)
	effectstest.PassPriorityAroundTable(t, restored)
	offer := learnOffer(restored, rme)
	if offer == nil {
		t.Fatalf("the restored table did not offer the Phoenix: %+v", restored.PendingChoices)
	}
	effectstest.AnswerOptional(t, restored, rme, true)
	effectstest.PassPriorityAroundTable(t, restored)
	if len(restored.PendingChoices) != 0 || rme.Graveyard.Size() != 1 {
		// Cram Session itself is the one card in the graveyard.
		t.Errorf("restored: graveyard %d, prompts %+v", rme.Graveyard.Size(), restored.PendingChoices)
	}
}
