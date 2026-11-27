package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Collective Restraint — Enchantment {3}{U} (Invasion, #5616):
//
//	"Domain — Creatures can't attack you unless their controller pays {X} for
//	 each creature they control that's attacking you, where X is the number of
//	 basic land types among lands you control."
//
// Propaganda's attack tax with a price counted off the board (Sphere of
// Safety's AttackTaxCounting): X is the defender's domain as each attack is
// declared, and a domain of zero is a tax of nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51f561db-ca7f-4661-bc04-99b8af02b8c9",
		Name:         "Collective Restraint",
		Completeness: effects.CompletenessFull,
		AttackTaxes: []game.AttackTax{
			effects.AttackTaxCounting(func(q game.AttackTaxQuery) int { return effects.DomainCount(q.Game, q.Defender) },
				"Domain — Creatures can't attack you unless their controller pays {X} for each creature they control that's attacking you, where X is the number of basic land types among lands you control."),
		},
	})
}
