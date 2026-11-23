package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Manticore — Creature — Manticore {3}{B}, 2/1 (AFR):
//
//	"Flash
//	 Flying
//	 Tail Spikes — When this creature enters, destroy target creature an
//	 opponent controls that was dealt damage this turn."
func init() {
	trig := effects.WhenThisEnters("Manticore — Tail Spikes: destroy target creature an opponent controls that was dealt damage this turn",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			target := effects.FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			return effects.DestroyTarget{Target: target}.Apply(ctx)
		})
	trig.Targets = effects.TargetCreature("target creature an opponent controls that was dealt damage this turn",
		effects.OpponentControls(), effects.DealtDamageThisTurn())
	effects.Register(effects.Spec{
		OracleID:        "fe35b55e-52ae-428e-b27e-817ee78b5d8c",
		Name:            "Manticore",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered:       []game.TriggeredAbility{trig},
	})
}
