package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gluttonous Guest — Creature — Vampire {2}{B}, 1/4 (Innistrad: Crimson
// Vow, #5495):
//
//	"When this creature enters, create a Blood token.
//	 Whenever you sacrifice a Blood token, you gain 1 life."
//
// The second ability fires once per Blood token sacrificed, including the
// one the Guest's own Blood spends as a cost (the sacrifice event is read
// last-known, CR 701.21).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d338aa5-22bc-4cc4-a1e0-cc5018ce001e",
		Name:         "Gluttonous Guest",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Gluttonous Guest — create a Blood token",
				effects.Do(effects.CreateToken{Template: effects.BloodToken(), N: 1})),
			effects.On(game.EventSacrifice, effects.VowB04YouSacrificedBlood, "Gluttonous Guest — gain 1 life",
				effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
