package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bat'leth — Artifact — Equipment {2}{B} (Star Trek):
//
//	When this Equipment enters, target creature an opponent controls gets
//	-2/-2 until end of turn.
//	Equipped creature gets +2/+0.
//	Equip {2}
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1000d26e-7a29-48df-806f-9eb7d4e1082d",
		Name:         "Bat'leth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Bat'leth — target creature an opponent controls gets -2/-2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: id, Power: -2, Toughness: -2, Label: "Bat'leth — -2/-2 until end of turn"}.Apply(ctx)
				}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.HarmUntilEOT(0, -2, -2))),
		},
		Static:    []game.StaticAbility{effects.PumpAttached(2, 0)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
