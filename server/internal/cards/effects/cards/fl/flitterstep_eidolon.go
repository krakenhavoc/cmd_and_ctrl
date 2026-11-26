package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flitterstep Eidolon — Enchantment Creature — Spirit {1}{U}, 1/1 (Born of the
// Gods, S82 OLD b377, #5559):
//
//	"Bestow {5}{U}
//	 This creature can't be blocked.
//	 Enchanted creature gets +1/+1 and can't be blocked."
//
// Invisible Stalker's RestrictSelf for the creature, RestrictAttached for the
// host (a bestowed Aura is no creature, so its own clause is inert).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "4280191f-e2cb-4cee-a3b4-660b3efd62e9",
		Name:             "Flitterstep Eidolon",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Bestow("{5}{U}")},
		Static: []game.StaticAbility{
			effects.RestrictSelf(game.CantBeBlocked),
			effects.PumpAttached(1, 1),
			effects.RestrictAttached(game.CantBeBlocked),
		},
	})
}
