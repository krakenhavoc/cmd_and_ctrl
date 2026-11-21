package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crushing Vines — Instant {2}{G} (Dark Ascension, #5571):
//
//	"Choose one —
//	 • Destroy target creature with flying.
//	 • Destroy target artifact."
//
// Crushing Canopy with an artifact bullet. Only the chosen bullet's clause
// is offered (ADR 0065), and a target gone by resolution does nothing (CR
// 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aebe0ba5-70b2-406b-b53b-03b12a95c2c2",
		Name:         "Crushing Vines",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: effects.ChooseOne(
			effects.Mode("Destroy target creature with flying.", effects.TargetCreature("target creature with flying", effects.HasKeyword("flying"))),
			effects.Mode("Destroy target artifact.", effects.TargetPermanent("target artifact", effects.Artifact())),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyFirstLegalTarget(ctx.Game, item)
		},
	})
}
