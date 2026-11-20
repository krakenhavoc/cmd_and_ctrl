package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chromanticore — Enchantment Creature — Manticore {W}{U}{B}{R}{G}, 4/4:
//
//	"Bestow {2}{W}{U}{B}{R}{G}
//	 Flying, first strike, vigilance, trample, lifelink
//	 Enchanted creature gets +4/+4 and has flying, first strike,
//	 vigilance, trample, and lifelink."
//
// Bestow is the shared alternative cost (CR 702.103); the engine makes
// the bestowed spell an Aura, attaches it and keeps it as a creature when
// it falls off. The two "enchanted creature" lines reach nothing while
// the card is a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "695ee4fe-356d-47fb-9442-d171b618151c",
		Name:            "Chromanticore",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", "vigilance", "trample", "lifelink"},
		AlternativeCosts: []game.AlternativeCost{
			effects.Bestow("{2}{W}{U}{B}{R}{G}"),
		},
		Static: []game.StaticAbility{
			effects.PumpAttached(4, 4),
			effects.GrantToAttached("flying", "first strike", "vigilance", "trample", "lifelink"),
		},
	})
}
