package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frenzied Baloth — Creature — Beast {G}{G}, 3/2:
//
//	"This spell can't be countered.
//	 Trample, haste
//	 Creature spells you control can't be countered.
//	 Combat damage can't be prevented."
//
// The counter riders are ADR 0106 §4's (#1806); "Combat damage can't be
// prevented" is ADR 0107 §5's battlefield static, which covers every
// player's combat damage, not only the Baloth's controller's.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6a4ef075-9254-4d36-9572-622833fae54e",
		Name:            "Frenzied Baloth",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"trample", "haste"},
		SpellsCantBeCountered: []game.CounterShieldStatic{
			effects.SpellsYouControlCantBeCountered("Creature spells you control can't be countered.", effects.Creature()),
		},
		DamageCantBePrevented: effects.CombatDamageCantBePrevented(),
	})
}
