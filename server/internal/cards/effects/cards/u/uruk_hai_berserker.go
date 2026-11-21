package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uruk-hai Berserker — Creature — Orc Berserker {2}{B}, 3/2:
//
//	"When this creature enters, the Ring tempts you."
//
// The tempt is the keyword action (ADR 0114, CR 701.54). The creature
// is on the battlefield as the trigger resolves, so it can be chosen as
// the Ring-bearer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e353a41-b523-4c22-9bda-d10460ba99c7",
		Name:         "Uruk-hai Berserker",
		Purpose:      game.Purpose{RingTempts: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Uruk-hai Berserker — the Ring tempts you", effects.Do(effects.TheRingTemptsYou{})),
		},
	})
}
