package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Valor Singer — Creature — Tiefling Bard {2}{R}, 2/3 (AFR):
//
//	"Combat Inspiration — At the beginning of combat on your turn, target
//	 creature you control gets +1/+0 until end of turn."
func init() {
	label := "Valor Singer — Combat Inspiration: target creature you control gets +1/+0 until end of turn"
	trig := effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true), label,
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			target := effects.FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			return effects.BoostUntilEOT{Target: target, Power: 1, Label: label}.Apply(ctx)
		})
	trig.Targets = effects.TargetCreature("target creature you control", effects.YouControl())
	effects.Register(effects.Spec{
		OracleID:     "3e0d39e2-d99f-4729-899b-c7bcd2b53615",
		Name:         "Valor Singer",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trig},
	})
}
