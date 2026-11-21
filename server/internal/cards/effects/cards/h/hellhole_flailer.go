package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hellhole Flailer — Creature — Ogre Warrior {1}{B}{R}, 3/2:
//
//	"Unleash (You may have this creature enter with a +1/+1 counter on
//	 it. It can't block as long as it has a +1/+1 counter on it.)
//	 {2}{B}{R}, Sacrifice this creature: It deals damage equal to its
//	 power to target player or planeswalker."
//
// Unleash is the engine's keyword (ADR 0109 §10). The sacrifice is the
// cost, so by resolution the Flailer is gone: "its power" is its power
// as it last existed on the battlefield (CR 113.7a, ctx.SourcePermanent),
// an unleash counter included, and it is still the damage's source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e13c1703-1184-4a8c-9310-f11e4e673597",
		Name:            "Hellhole Flailer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordUnleash},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}{R}, Sacrifice this creature: It deals damage equal to its power to target player or planeswalker.",
			Cost:    effects.Plus(effects.ManaCost("{2}{B}{R}"), effects.SacrificeThis()),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Effect:  effects.SacrificedSourceDealsItsPowerToTheTarget,
		}},
	})
}
