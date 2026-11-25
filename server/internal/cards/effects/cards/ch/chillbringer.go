package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chillbringer — Creature — Elemental {4}{U}, 3/3 (Ravnica Allegiance):
//
//	"Flying
//	 When this creature enters, tap target creature an opponent controls. It
//	 doesn't untap during its controller's next untap step."
//
// Rooftop Nuisance's tap-and-freeze, aimed at an opponent's creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "604d2a0e-224c-4a7b-aabf-033e7dcee3f2",
		Name:            "Chillbringer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Chillbringer — tap target creature an opponent controls; it doesn't untap during its controller's next untap step",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Chillbringer — it doesn't untap during its controller's next untap step"}.Apply(ctx)
				}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
