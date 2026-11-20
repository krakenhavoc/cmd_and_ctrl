package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Robaran Mercenaries — Creature — Human Mercenary {3}{W}, 3/4:
//
//	"Vigilance
//	 This creature has all activated abilities of all legendary
//	 creatures you control."
//
// A layer-6 borrow over the battlefield (ADR 0155). Legendary is read off
// the layered type line, so a creature made legendary lends its rows and
// one that stops being legendary stops.
//
// No simplification.
func init() {
	const clause = "This creature has all activated abilities of all legendary creatures you control."
	effects.Register(effects.Spec{
		OracleID:        "1d8c8c71-d569-450d-adc1-1a1b9aea0123",
		Name:            "Robaran Mercenaries",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{effects.HasActivatedAbilitiesOf(game.AbilityBorrow{
			Label:   clause,
			Lenders: effects.BattlefieldLenders(effects.ByController(effects.And(effects.Creature(), effects.Legendary(), effects.YouControl()))),
		})},
	})
}
