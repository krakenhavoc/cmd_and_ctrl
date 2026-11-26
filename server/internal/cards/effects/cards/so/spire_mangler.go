package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spire Mangler — Creature — Insect {2}{B}, 2/1 (Ravnica Allegiance):
//
//	"Flash
//	 Flying
//	 When this creature enters, target creature you control with flying gets
//	 +2/+0 until end of turn."
//
// The target is chosen as the trigger is put on the stack, so the Mangler
// itself (it has flying) is a legal pick.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6416dae0-0d40-4d88-b30b-e1a0c8380e78",
		Name:            "Spire Mangler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Spire Mangler — target creature you control with flying gets +2/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: id, Power: 2, Label: "Spire Mangler — +2/+0 until end of turn"}.Apply(ctx)
				}),
				effects.TargetCreature("target creature you control with flying", effects.YouControl(), effects.HasKeyword("flying"))),
		},
	})
}
