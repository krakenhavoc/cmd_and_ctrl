package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bat'leth — Artifact — Equipment {2}{B} (Star Trek):
//
//	When this Equipment enters, target creature an opponent controls gets
//	-2/-2 until end of turn.
//	Equipped creature gets +2/+0.
//	Equip {2}
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1000d26e-7a29-48df-806f-9eb7d4e1082d",
		Name:         "Bat'leth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Bat'leth — target creature an opponent controls gets -2/-2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return BoostUntilEOT{Target: id, Power: -2, Toughness: -2, Label: "Bat'leth — -2/-2 until end of turn"}.Apply(ctx)
				}),
				TargetCreature("target creature an opponent controls", OpponentControls())),
				ForTargets(HarmUntilEOT(0, -2, -2))),
		},
		Static:    []game.StaticAbility{PumpAttached(2, 0)},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
