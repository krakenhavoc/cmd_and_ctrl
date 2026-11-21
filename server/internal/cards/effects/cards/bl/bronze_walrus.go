package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bronze Walrus — Artifact Creature — Walrus {3}, 2/2 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"When this creature enters, scry 2.
//	 {T}: Add one mana of any color."
//
// Myr Custodian's scry 2 on enters and the ordinary any-colour tap ability
// (CR 605.1a). The Walrus is summoning sick like any creature, so the mana
// ability needs it to have been under your control since the turn began
// (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3fdc76b0-450a-4fcb-8a64-113edce67f3d",
		Name:         "Bronze Walrus",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bronze Walrus — scry 2", effects.Do(effects.Scry{N: 2})),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color.",
		}},
	})
}
