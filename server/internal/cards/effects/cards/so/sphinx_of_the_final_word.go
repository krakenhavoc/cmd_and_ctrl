package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphinx of the Final Word — Creature — Sphinx {5}{U}{U}, 5/5:
//
//	"This spell can't be countered.
//	 Flying
//	 Hexproof
//	 Instant and sorcery spells you control can't be countered."
//
// Flying and hexproof are keywords; the rider is Spec.CantBeCountered;
// the last line is ADR 0106 §4's battlefield static (#1806).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d4246e4d-390d-4925-a5a8-89cd096a237c",
		Name:            "Sphinx of the Final Word",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "hexproof"},
		CantBeCountered: true,
		SpellsCantBeCountered: []game.CounterShieldStatic{
			effects.SpellsYouControlCantBeCountered("Instant and sorcery spells you control can't be countered.", effects.Or(effects.Instant(), effects.Sorcery())),
		},
	})
}
