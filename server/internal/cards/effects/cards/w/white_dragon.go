package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// White Dragon — Creature — Dragon {4}{W}{W}, 4/4 (Adventures in the Forgotten Realms):
//
//	Flying
//	 Cold Breath — When this creature enters, tap target creature an opponent controls. That creature doesn't untap during its controller's next untap step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bb03e682-4525-40d2-8c49-6328f84bd36a",
		Name:            "White Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("White Dragon — Cold Breath: tap target creature an opponent controls, it doesn't untap", func(g *game.Game, item *game.StackItem) error {
				return effects.B751TapFreezeTargets(effects.NewContext(g, item), uuid.Nil, "White Dragon")
			}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
