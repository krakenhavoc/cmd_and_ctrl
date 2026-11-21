package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hard-Won Jitte — {1}{R} Artifact — Equipment (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"Equipped creature has double strike.
//	 Equip {2}"
//
// Fireshrieker's text at a cheaper casting cost: a layer-6 grant to the
// equipped creature and the ordinary equip ability (CR 702.6a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab433d41-e801-405c-81f7-06fa6640797d",
		Name:         "Hard-Won Jitte",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.GrantToAttached("double strike"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}
