package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sek'Kuar, Deathkeeper — Legendary Creature — Orc Shaman {2}{B}{R}{G},
// 4/3 (Commander Masters):
//
//	"Whenever another nontoken creature you control dies, create a 3/1
//	 black and red Graveborn creature token with haste."
//
// "You control" and "nontoken" both look at the creature as it died
// (CR 603.10a); one token per death.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94426127-65c2-435e-ba92-423a3c102061",
		Name:         "Sek'Kuar, Deathkeeper",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventLTB, effects.AllOf(effects.AnotherCreatureDied, effects.ACreatureYouControlDied, gapB101NontokenCreatureDied),
				"Sek'Kuar, Deathkeeper — create a 3/1 black and red Graveborn creature token with haste",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/1 black and red Graveborn with haste"), N: 1})), game.Purpose{DeathPayoff: true}),
		},
	})
}

// gapB101NontokenCreatureDied is "a nontoken creature died" (any controller).
func gapB101NontokenCreatureDied(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := effects.DiedCreature(ev, g)
	return ok && !effects.IsToken(dead)
}
