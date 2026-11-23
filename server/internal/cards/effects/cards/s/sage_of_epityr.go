package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sage of Epityr — Creature — Human Wizard {U}, 1/1 (Time Spiral, #5596):
//
//	"When this creature enters, look at the top four cards of your
//	 library, then put them back in any order."
//
// Sage Owl's trigger without the wings: LookAtTop over four cards, put
// back in the order its controller picks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef893585-246c-4028-bd2b-14ac51ce29be",
		Name:         "Sage of Epityr",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sage of Epityr — look at the top four cards of your library, then put them back in any order", effects.Do(effects.LookAtTop{N: 4})),
		},
	})
}
