package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "25977c02-e2d4-4afd-b12d-37ce4d58f453",
		Name:         "Golden Egg",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{samiDrawOnETB("Golden Egg")},
		ManaAbilities: []ManaAbility{
			samiAnyColorMana(ManaAbilityCost{Tap: true, Sacrifice: true, Mana: "{1}"},
				"{1}, {T}, Sacrifice this artifact: Add one mana of any color"),
		},
		Activated: []ActivatedAbility{samiFoodLifeAbility()},
	})
}
