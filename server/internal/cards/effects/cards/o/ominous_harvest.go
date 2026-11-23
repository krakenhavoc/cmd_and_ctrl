package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ominous Harvest — Sorcery {2}{B} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Gravestorm (When you cast this spell, copy it for each permanent put
//	 into a graveyard from the battlefield this turn.)
//	 Target player draws a card and loses 1 life."
//
// Gravestorm is CR 702.69a, soc_b9_helpers.go's Gravestorm(): storm's
// shape with the turn's battlefield-to-graveyard count, every player's
// permanents and tokens included (the rulings). Each copy may choose a
// new target player (CR 707.10c). The spell is Ancestral Craving's body
// with one card and one life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76bc6518-dea4-4a5c-a8a7-352cd297197b",
		Name:         "Ominous Harvest",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 1, LifeLoss: 1}),
		Triggered:    []game.TriggeredAbility{effects.Gravestorm()},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B32TargetPlayerDrawsAndLosesLife(item, ctx, 1, 1)
		},
	})
}
