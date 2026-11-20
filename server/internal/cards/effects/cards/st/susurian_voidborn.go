package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Susurian Voidborn — Creature — Vampire Soldier {2}{B}, 2/2 (Edge of
// Eternities, #4461):
//
//	"Whenever this creature or another creature or artifact you control
//	 dies, target opponent loses 1 life and you gain 1 life.
//	 Warp {B}"
//
// Al Bhed Salvagers' trigger, word for word: the Voidborn's own death, or
// another creature or artifact its controller controlled being put into a
// graveyard from the battlefield (CR 700.4), read from the departure
// (CR 603.10a) so a crewed Vehicle counts as a creature and a token
// counts. Warp is the shared alternative cost (CR 702.185). A warped
// Voidborn is exiled at the end step, not put into a graveyard, so that
// exile drains nobody.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "4f754f11-3b6d-4a4e-86d7-7495e475ef74",
		Name:             "Susurian Voidborn",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Warp("{B}")},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.ThisDied(ev, source, lki, g) || effects.AnotherCreatureOrArtifactYouControlWasPutIntoAGraveyard(ev, source, lki, g)
			},
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Key:     "Susurian Voidborn — target opponent loses 1 life and you gain 1 life",
			Purpose: game.Purpose{DeathPayoff: true},
			Effect:  effects.Do(effects.TargetPlayerLosesLife{N: 1}, effects.GainLife{Amount: 1}),
		}},
	})
}
