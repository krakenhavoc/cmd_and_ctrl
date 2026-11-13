package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chainer's Edict — Sorcery {1}{B} (Torment, #5612):
//
//	"Target player sacrifices a creature of their choice.
//	 Flashback {5}{B}{B} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// The player is the target and the creature is their own pick at
// resolution, so hexproof or protection on their creatures doesn't
// matter (CR 701.21a): the shared targetPlayerSacrificesACreature body
// Crude Bent Blade and Scarring Memories use, which does nothing when
// the player is no longer a legal target (CR 608.2b). Flashback is the
// shared constructor: the graveyard cast, the price, and CR 702.34a's
// exile as the spell leaves the stack.
//
// No simplification.
func init() {
	const name = "Chainer's Edict"
	Register(Spec{
		OracleID:         "7975cb9d-d37e-44f8-9614-86d2ad453c35",
		Name:             name,
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		Targets:          TargetPlayer("target player"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{B}{B}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return targetPlayerSacrificesACreature(name+" — sacrifice a creature")(ctx.Game, item)
		},
	})
}
