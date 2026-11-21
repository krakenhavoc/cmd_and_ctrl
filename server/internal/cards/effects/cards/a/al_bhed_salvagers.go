package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Al Bhed Salvagers — Creature — Human Artificer Warrior {2}{B}, 2/3
// (Final Fantasy, #4462):
//
//	"Whenever this creature or another creature or artifact you control
//	 dies, target opponent loses 1 life and you gain 1 life."
//
// One trigger, two conditions: the Salvagers' own death, and another
// creature or artifact its controller controlled being put into a
// graveyard from the battlefield (CR 700.4), read from the departure
// (CR 603.10a) so a crewed Vehicle counts as a creature and a token
// counts. The drain is Blood Artist's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43f81514-e812-4e17-9b1a-704c136d8ffd",
		Name:         "Al Bhed Salvagers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.ThisDied(ev, source, lki, g) || effects.AnotherCreatureOrArtifactYouControlWasPutIntoAGraveyard(ev, source, lki, g)
			},
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Key:     "Al Bhed Salvagers — target opponent loses 1 life and you gain 1 life",
			Purpose: game.Purpose{DeathPayoff: true},
			Effect:  effects.Do(effects.TargetPlayerLosesLife{N: 1}, effects.GainLife{Amount: 1}),
		}},
	})
}
