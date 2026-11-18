package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Manticore — Creature — Manticore {3}{B}, 2/1 (AFR):
//
//	"Flash
//	 Flying
//	 Tail Spikes — When this creature enters, destroy target creature an
//	 opponent controls that was dealt damage this turn."
func init() {
	trig := WhenThisEnters("Manticore — Tail Spikes: destroy target creature an opponent controls that was dealt damage this turn",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			target := FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			return DestroyTarget{Target: target}.Apply(ctx)
		})
	trig.Targets = TargetCreature("target creature an opponent controls that was dealt damage this turn",
		OpponentControls(), DealtDamageThisTurn())
	Register(Spec{
		OracleID:        "fe35b55e-52ae-428e-b27e-817ee78b5d8c",
		Name:            "Manticore",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered:       []game.TriggeredAbility{trig},
	})
}
