package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Golden Egg — Artifact — Food {2}:
//
//	"When this artifact enters, draw a card.
//	 {1}, {T}, Sacrifice this artifact: Add one mana of any color.
//	 {2}, {T}, Sacrifice this artifact: You gain 3 life."
//
// The mana half is a mana ability (no stack); the life half is the
// Food ability on the stack (samiFoodLifeAbility). The Food subtype
// comes off the type line, so Food payoffs see it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25977c02-e2d4-4afd-b12d-37ce4d58f453",
		Name:         "Golden Egg",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.SamiDrawOnETB("Golden Egg")},
		ManaAbilities: []effects.ManaAbility{
			effects.SamiAnyColorMana(effects.ManaAbilityCost{Tap: true, Sacrifice: true, Mana: "{1}"},
				"{1}, {T}, Sacrifice this artifact: Add one mana of any color"),
		},
		Activated: []effects.ActivatedAbility{effects.SamiFoodLifeAbility()},
	})
}
