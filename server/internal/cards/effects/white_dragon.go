package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// White Dragon — Creature — Dragon {4}{W}{W}, 4/4 (Adventures in the Forgotten Realms):
//
//	Flying
//	 Cold Breath — When this creature enters, tap target creature an opponent controls. That creature doesn't untap during its controller's next untap step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bb03e682-4525-40d2-8c49-6328f84bd36a",
		Name:            "White Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("White Dragon — Cold Breath: tap target creature an opponent controls, it doesn't untap", func(g *game.Game, item *game.StackItem) error {
				return b751TapFreezeTargets(NewContext(g, item), uuid.Nil, "White Dragon")
			}),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
