package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warden of the Wall — Artifact {3}:
//
//	"This artifact enters tapped.
//	 {T}: Add {C}.
//	 During turns other than yours, this artifact is a 2/3 Gargoyle artifact
//	 creature with flying."
//
// Three statics under one condition read live from whose turn it is: the
// Creature type and Gargoyle subtype (layer 4), 2/3 (layer 7b), flying (layer 6).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5d66c2b9-6011-4168-b468-0bde50bac37e",
		Name:         "Warden of the Wall",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Static: effects.OldB443GargoyleOnOthersTurns(),
	})
}
