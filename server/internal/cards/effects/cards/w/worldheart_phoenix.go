package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Worldheart Phoenix — Creature — Phoenix {3}{R}, 2/2 (Conflux):
//
//	"Flying
//	 You may cast this card from your graveyard by paying {W}{U}{B}{R}{G}
//	 rather than paying its mana cost. If you do, it enters with two +1/+1
//	 counters on it."
//
// The graveyard is opened as a cast zone and the only way in from it is the
// alternative cost, as Squee, Dubious Monarch's; the counters hang off that
// cost (an escape's shape, AlternativeCost.EntersWithCounterName), so they
// ride the entry event under doublers (CR 614) and only a Phoenix cast this
// way gets them, never one reanimated or hard-cast from the hand. The
// alternative cost has no exile clause, so a countered Phoenix is back in the
// graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "67c50af9-bb8e-40b6-8b25-c209784b6865",
		Name:            "Worldheart Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CastableZones:   []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{{
			Key:                    "worldheart-phoenix-graveyard",
			Label:                  "Cast from your graveyard by paying {W}{U}{B}{R}{G} rather than paying its mana cost",
			ManaCost:               "{W}{U}{B}{R}{G}",
			FromZone:               game.ZoneGraveyard,
			EntersWithCounterName:  game.CounterPlusOne,
			EntersWithCounterCount: 2,
		}},
	})
}
