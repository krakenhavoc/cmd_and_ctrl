package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Temple Acolyte — Creature — Human Cleric {1}{W}, 1/3 (Portal Second
// Age, #5625):
//
//	"When this creature enters, you gain 3 life."
//
// Hill Giant Herdgorger's entry trigger on a 1/3: it uses the stack, so
// it can be answered, and the life is gained as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17e4085e-af44-4ba8-af05-ccfe4b35f96a",
		Name:         "Temple Acolyte",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Temple Acolyte — gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
