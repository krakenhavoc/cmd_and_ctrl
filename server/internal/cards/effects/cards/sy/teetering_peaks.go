package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teetering Peaks — Land (Zendikar, #5583):
//
//	"This land enters tapped.
//	 When this land enters, target creature gets +2/+0 until end of turn.
//	 {T}: Add {R}."
//
// Sandstone Bridge's shape with a +2/+0 boost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a9437a6-4e61-48b6-8194-1c6ba6432250",
		Name:         "Teetering Peaks",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Teetering Peaks — target creature gets +2/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: t.ID, Power: 2, Label: "Teetering Peaks — +2/+0"}.Apply(ctx)
				}), effects.TargetCreature("target creature")),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
	})
}
