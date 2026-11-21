package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Holy Cow — Creature — Ox Angel {2}{W}, 2/2 (Outlaws of Thunder
// Junction, #4468):
//
//	"Flash
//	 Flying
//	 When this creature enters, you gain 2 life and scry 1."
//
// Flash and flying are printed keywords. The enters trigger is
// Compassionate Healer's two footprinted Do steps, gain then scry, in
// the printed order (CR 701.22a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "748003fd-84c5-4c99-b146-d7d382eef64d",
		Name:            "Holy Cow",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Purpose:         game.Purpose{LifeGain: 2, Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Holy Cow — you gain 2 life and scry 1",
				effects.Do(effects.GainLife{Amount: 2}, effects.Scry{N: 1})),
		},
	})
}
