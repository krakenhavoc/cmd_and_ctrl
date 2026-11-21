package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tocasia's Onulet — Artifact Creature — Construct {5}, 4/4:
//
//	"When this creature leaves the battlefield, you gain 2 life.
//	 Unearth {3}{W}"
//
// An unearthed Onulet that is exiled at end step leaves the battlefield
// and triggers again (CR 603.10a).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "150dd67b-4de2-4fcd-b753-e4096de911a2",
		Name:         "Tocasia's Onulet",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisLeaves("Tocasia's Onulet — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{3}{W}")},
	})
}
