package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mishra's Self-Replicator — Artifact Creature — Assembly-Worker {5}, 2/2
// (Dominaria, S82 OLD batch 215, #5530):
//
//	"Whenever you cast a historic spell, you may pay {1}. If you do, create a
//	 token that's a copy of this creature. (Artifacts, legendaries, and Sagas
//	 are historic.)"
//
// Jamie McCrimmon's cast trigger (only your own casts, historic read on the
// spell on the stack) with Bitter Chill's "may pay {1}" asked as it resolves
// (CR 603.5, 118.12). A yes is Myr Propagator's token copy of this creature
// (CR 707.2): its copiable values as it is, or as it last existed if it has
// left. The token has the same trigger, so it copies again on the next
// historic spell.
//
// No simplification.
func init() {
	const label = "Mishra's Self-Replicator — you may pay {1}; if you do, create a token that's a copy of this creature"
	effects.Register(effects.Spec{
		OracleID:     "5546d127-a961-4e33-bbc9-06ee14b0cb04",
		Name:         "Mishra's Self-Replicator",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.B14Historic(), label, func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}",
					Question: "Mishra's Self-Replicator — pay {1} to create a token that's a copy of this creature?",
					Then:     mishrasSelfReplicatorPaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// mishrasSelfReplicatorPaid is the paid branch: the token copy of the source. An
// on-disk key: never renamed, never reused.
var mishrasSelfReplicatorPaid = effects.PayUnlessBranches("pay-unless/mishras-self-replicator-copy",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateTokenCopy{Controller: ctx.Controller(), Copy: ctx.Source(), N: 1}.Apply(ctx)
	}, nil)
