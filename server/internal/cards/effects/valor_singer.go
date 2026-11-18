package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Valor Singer — Creature — Tiefling Bard {2}{R}, 2/3 (AFR):
//
//	"Combat Inspiration — At the beginning of combat on your turn, target
//	 creature you control gets +1/+0 until end of turn."
func init() {
	label := "Valor Singer — Combat Inspiration: target creature you control gets +1/+0 until end of turn"
	trig := On(game.EventStepBegan, StepBegan(game.StepBeginCombat, true), label,
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			target := FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			return BoostUntilEOT{Target: target, Power: 1, Label: label}.Apply(ctx)
		})
	trig.Targets = TargetCreature("target creature you control", YouControl())
	Register(Spec{
		OracleID:     "3e0d39e2-d99f-4729-899b-c7bcd2b53615",
		Name:         "Valor Singer",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trig},
	})
}
