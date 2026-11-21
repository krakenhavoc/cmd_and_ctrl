package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inherited Envelope — Artifact {3}:
//
//	"When this artifact enters, the Ring tempts you.
//	 {T}: Add one mana of any color."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f1ae834-f6b8-458c-ac0d-3510ddb85193",
		Name:         "Inherited Envelope",
		Purpose:      game.Purpose{RingTempts: 1},
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Inherited Envelope — the Ring tempts you", effects.Do(effects.TheRingTemptsYou{})),
		},
	})
}
