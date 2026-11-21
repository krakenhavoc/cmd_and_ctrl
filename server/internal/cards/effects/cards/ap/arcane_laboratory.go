package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcane Laboratory — Enchantment {2}{U} (Urza's Saga):
//
//	"Each player can't cast more than one spell each turn."
//
// Rule of Law's restriction under another name: the cast tally the
// engine keeps per player per turn, read while this permanent is on the
// battlefield, so it lifts the moment Arcane Laboratory leaves. "Each
// player" includes its controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5e3c26c-c28c-455f-a682-2597e4b92d10",
		Name:         "Arcane Laboratory",
		Completeness: effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			effects.EachPlayerMaxSpellsPerTurn(1,
				"Arcane Laboratory — each player can't cast more than one spell each turn."),
		},
	})
}
