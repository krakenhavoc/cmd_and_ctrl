package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Necra Disciple — Creature — Human Wizard {B}, 1/1 (Apocalypse, S82 OLD,
// #5614):
//
//	"{G}, {T}: Add one mana of any color.
//	 {W}, {T}: Prevent the next 1 damage that would be dealt to any target this
//	 turn."
//
// A mana ability with a mana cost and a one-damage shield (CR 302.6 applies to
// both).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "ef85a5eb-9cde-47b7-97e4-daa9aa9d4676",
		Name:          "Necra Disciple",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.OldB632DiscipleAnyColorMana("{G}")},
		Activated:     []effects.ActivatedAbility{effects.OldB632PreventOneRow("Necra Disciple", "{W}")},
	})
}
