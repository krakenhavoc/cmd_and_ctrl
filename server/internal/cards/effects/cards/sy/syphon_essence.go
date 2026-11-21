package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Syphon Essence — Instant {2}{U} (Innistrad: Crimson Vow):
//
//	"Counter target creature or planeswalker spell. Create a Blood token."
//
// If the spell is gone on resolution the target is illegal and this does
// nothing, Blood token included (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6e7b5ba-ea10-4f53-8afc-045aa1906c6b",
		Name:         "Syphon Essence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target creature or planeswalker spell", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			if err := (effects.CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.S82Vow03CreateBlood(1).Apply(ctx)
		},
	})
}
