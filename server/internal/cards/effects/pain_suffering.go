package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pain // Suffering — split card (Invasion, #5616):
//
//	Pain — Sorcery {B}: "Target player discards a card."
//	Suffering — Sorcery {3}{R}: "Destroy target land."
//
// Each half is an ADR 0034 face (ADR 0103): Pain under the bare oracle
// ID, Suffering under "#1". Only one half can be cast (CR 709.3; the card
// has no fuse). Pain is Mind Peel's targetOpponentDiscards(1) on any
// player, the caster included; the discarding player picks the card (CR
// 701.9b). Suffering is Sinkhole's destroyTheTargetPermanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "433fcba3-06b8-4f1a-a709-300838251439",
		Name:         "Pain",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return targetOpponentDiscards(1)(ctx.Game, item)
		},
	})
	Register(Spec{
		OracleID:     "433fcba3-06b8-4f1a-a709-300838251439#1",
		Name:         "Suffering",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target land", Land()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
