package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smoldering Spires — Land  (Worldwake, #5582):
//
//	"This land enters tapped.
//	 When this land enters, target creature can't block this turn.
//	 {T}: Add {R}."
//
// Halimar Depths' shape: the tapped entry is a replacement, the
// enters trigger uses the stack and takes Ahn-Crop Crasher's
// "target creature can't block this turn".
//
// No simplification.
const smolderingSpiresLabel = "Smoldering Spires — target creature can't block this turn"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "cfa3288d-e521-4a13-bcb3-7950a94e1746",
		Name:         "Smoldering Spires",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters(smolderingSpiresLabel, effects.TargetCreatureCantBlockThisTurn(smolderingSpiresLabel)),
				effects.TargetCreature("target creature")),
		},
		ManaAbilities: []effects.ManaAbility{effects.Cmdr2B01TapForColor("R")},
	})
}
