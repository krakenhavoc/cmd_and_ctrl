package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grim Bounty — Sorcery {2}{B}{B} (Adventures in the Forgotten Realms):
//
//	"Destroy target creature or planeswalker. Create a Treasure token."
//
// The Treasure is made whether or not the destroy happens (an indestructible
// target is still a legal target); only an illegal target at resolution
// fizzles the whole spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8060dc97-9868-4846-85f4-0ef70da2b021",
		Name:         "Grim Bounty",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (effects.DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
		},
	})
}
