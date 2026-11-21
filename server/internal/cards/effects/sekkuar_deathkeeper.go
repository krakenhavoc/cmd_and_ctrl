package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "94426127-65c2-435e-ba92-423a3c102061",
		Name:         "Sek'Kuar, Deathkeeper",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventLTB, AllOf(AnotherCreatureDied, ACreatureYouControlDied, gapB101NontokenCreatureDied),
				"Sek'Kuar, Deathkeeper — create a 3/1 black and red Graveborn creature token with haste",
				Do(CreateToken{Template: TokenCard("3/1 black and red Graveborn with haste"), N: 1})), game.Purpose{DeathPayoff: true}),
		},
	})
}

// gapB101NontokenCreatureDied is "a nontoken creature died" (any controller).
func gapB101NontokenCreatureDied(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && !IsToken(dead)
}
