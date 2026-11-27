package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ohran Yeti — Snow Creature — Yeti {3}{R}, 3/3 (Coldsnap):
//
//	"{2}{S}: Target snow creature gains first strike until end of turn."
//
// The target must be a snow creature as the ability is activated and again as
// it resolves (CR 608.2b). The Yeti itself is a snow creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e606cba7-7e3c-4834-b2cd-56e51cedccb7",
		Name:         "Ohran Yeti",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{2}{S}: Target snow creature gains first strike until end of turn.",
			Purpose: effects.ForTargets(game.TargetPurpose{
				Slot: 0, UntilEndOfTurn: game.TargetBenefit, Pump: &game.Pump{Keywords: []string{"first strike"}},
			}),
			Cost:    effects.ManaCost("{2}{S}"),
			Targets: effects.TargetCreature("target snow creature", effects.Snow()),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("first strike", "Ohran Yeti — first strike until end of turn"),
		}},
	})
}
