package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// V.A.T.S. — Instant {2}{B}{B} (Fallout, #6100):
//
//	"Split second (As long as this spell is on the stack, players can't
//	 cast spells or activate abilities that aren't mana abilities.)
//	 Choose any number of target creatures with equal toughness. Destroy
//	 the chosen creatures."
//
// "With equal toughness" is a rule over the chosen set (CR 601.2c), the
// sameness set rule keyed on toughness after the layers
// (WithEqualToughness, game.TargetShareToughness): every pick must share
// one toughness as the spell is cast. On resolution (CR 608.2b) the picks
// still on the battlefield are judged again together, so a creature whose
// toughness changed in response no longer has equal toughness with the
// others and is an illegal target, and so is every pick it now disagrees
// with; a pick that left the battlefield is simply gone and the rest are
// still destroyed. Split second keeps that window to mana abilities and
// triggered abilities (CR 702.61a). The legal picks are destroyed
// together, as one event.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d077d54c-6f1b-4795-a241-6073b0dc1dc8",
		Name:            "V.A.T.S.",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordSplitSecond},
		Targets:         effects.TargetCreature("any number of target creatures with equal toughness").WithCount(0, 0).AllShare(effects.WithEqualToughness()),
		Purpose:         game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var doomed []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					doomed = append(doomed, t.ID)
				}
			}
			if len(doomed) > 0 {
				ctx.Game.DestroyPermanentsForEffect(doomed)
			}
			return nil
		},
	})
}
