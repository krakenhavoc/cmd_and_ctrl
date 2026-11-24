package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Trickster-God's Heist — Enchantment — Saga {2}{U}{B} (Kaldheim,
// #5505):
//
//	"I — You may exchange control of two target creatures.
//	 II — You may exchange control of two target nonbasic, noncreature
//	 permanents that share a card type.
//	 III — Target player loses 3 life and you gain 3 life."
//
// Both exchanges name their two targets as the chapter goes on the stack
// (CR 603.3d), judged again as it resolves (CR 608.2b); chapter II's second
// target is judged against the first's card types (Burglar's Plot's
// relation). "You may" is asked as the chapter resolves, with the two
// permanents carried, and the exchange happens only on a yes and only while
// both are still there (CR 701.12). Chapter III is the Saga's controller
// gaining 3 life as the target loses 3.
//
// No simplification.
func init() {
	const name = "The Trickster-God's Heist"
	nonbasicNoncreature := effects.And(effects.Noncreature(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return !c.HasSupertype("basic")
	})
	effects.Register(effects.Spec{
		OracleID:     "78f99c66-cb50-49e3-ab2d-e7755818ffd7",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "you may exchange control of two target creatures"),
				effects.TargetCreature("two target creatures").WithCount(2, 2),
				func(g *game.Game, item *game.StackItem) error {
					return trickstersHeistAsk(g, item, "creatures")
				}),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "you may exchange control of two target nonbasic, noncreature permanents that share a card type"),
				effects.Clauses(
					effects.TargetPermanent("target nonbasic, noncreature permanent", nonbasicNoncreature),
					effects.Distinct(effects.SharesACardTypeWithTarget(
						effects.TargetPermanent("another target nonbasic, noncreature permanent that shares a card type with it", nonbasicNoncreature), 0)),
				),
				func(g *game.Game, item *game.StackItem) error {
					return trickstersHeistAsk(g, item, "permanents")
				}),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "target player loses 3 life and you gain 3 life"),
				effects.TargetPlayer("target player"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.TargetPlayerLosesLifeYouGain(3)(item, effects.NewContext(g, item))
				}),
		},
	})
}

// trickstersHeistAsk asks "you may exchange control" over the chapter's two
// still-legal targets, carried to the keyed answer. Fewer than two legal
// targets is no exchange (CR 608.2b).
func trickstersHeistAsk(g *game.Game, item *game.StackItem, what string) error {
	ctx := effects.NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) != 2 {
		return nil
	}
	return effects.MayChoice{
		Question: "The Trickster-God's Heist — exchange control of the two " + what + "?",
		Then:     trickstersHeistExchange,
		Carry:    ids,
	}.Apply(ctx)
}

// trickstersHeistExchange is the yes: exchange control of the two carried
// permanents. An on-disk key.
var trickstersHeistExchange = effects.MayYes("confirm/the-trickster-gods-heist-exchange", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if len(r.Carry) != 2 {
		return nil
	}
	return effects.ExchangeControl{A: r.Carry[0], B: r.Carry[1], Label: "The Trickster-God's Heist — exchange control"}.Apply(ctx)
})
