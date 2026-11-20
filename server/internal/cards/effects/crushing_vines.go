package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "aebe0ba5-70b2-406b-b53b-03b12a95c2c2",
		Name:         "Crushing Vines",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: ChooseOne(
			Mode("Destroy target creature with flying.", TargetCreature("target creature with flying", HasKeyword("flying"))),
			Mode("Destroy target artifact.", TargetPermanent("target artifact", Artifact())),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyFirstLegalTarget(ctx.Game, item)
		},
	})
}
