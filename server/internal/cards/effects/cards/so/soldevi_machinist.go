package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soldevi Machinist — Creature — Human Wizard Artificer {1}{U}, 1/1 (Ice Age,
// S82 OLD batch 851, #5636):
//
//	"{T}: Add {C}{C}. Spend this mana only to activate abilities of artifacts."
//
// Myr Reservoir's two-colourless mana ability with its activate-only clause,
// narrowed to an artifact source: the mana pays for the activated ability of a
// permanent that is an artifact, and for nothing else (not a spell, not a
// creature's ability unless it is an artifact too).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06600efe-1718-4cc1-b1c0-59c82e2efd73",
		Name:         "Soldevi Machinist",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     "{C}{C}",
			Label:        "{T}: Add {C}{C}. Spend this mana only to activate abilities of artifacts.",
			Restrictions: []string{game.ManaRestrictActivate, game.ManaRestrictType("artifact")},
		}},
	})
}
