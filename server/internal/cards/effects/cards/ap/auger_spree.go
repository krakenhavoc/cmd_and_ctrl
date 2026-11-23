package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Auger Spree — Instant {1}{B}{R} (S82 GAP, #5948):
//
//	"Target creature gets +4/-4 until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0c96c7a-02d8-40cb-a08b-4a2a3b631eef",
		Name:         "Auger Spree",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			return effects.BoostUntilEOT{Target: item.Targets[0].ID, Power: 4, Toughness: -4, Label: "Auger Spree — +4/-4"}.Apply(ctx)
		},
	})
}
