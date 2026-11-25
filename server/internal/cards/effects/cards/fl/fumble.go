package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fumble — Instant {1}{U} (Battlebond):
//
//	"Return target creature to its owner's hand. Gain control of all Auras and
//	 Equipment that were attached to it, then attach them to another creature."
//
// Written as the printed steps (CR 608.2c). The Auras and Equipment are those
// attached when the spell resolves, read before the creature leaves. You gain
// control of each for good (CR 611.2a), then choose one other creature, any
// player's, and attach each of them that can legally be attached to it
// (CR 701.3b: an Aura's enchant clause and protection are checked); one that
// cannot stays unattached, and an Aura is then put into its owner's graveyard
// (CR 704.5m). A target that has left leaves the spell with nothing to do
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cbf68ddd-596f-4b2d-b93c-3b7976ba4d37",
		Name:         "Fumble",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.OldB206FumbleResolve(ctx.Game, item)
		},
	})
}
