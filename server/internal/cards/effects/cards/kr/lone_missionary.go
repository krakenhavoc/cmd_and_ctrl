package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lone Missionary — Creature — Kor Monk {1}{W}, 2/1 (Rise of the Eldrazi,
// #5581):
//
//	"When this creature enters, you gain 4 life."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "626f1dc3-3b14-4873-902f-ffec5487a97d",
		Name:         "Lone Missionary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Lone Missionary — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}
