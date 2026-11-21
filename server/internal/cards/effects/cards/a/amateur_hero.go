package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Amateur Hero — {2}{W} Creature — Human Hero, 3/3 (Marvel's Spider-Man
// Eternal, #4460):
//
//	"When this creature enters, you gain 2 life."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc8ee6e1-a2e1-4033-a164-7968fc79dd27",
		Name:         "Amateur Hero",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Amateur Hero — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
