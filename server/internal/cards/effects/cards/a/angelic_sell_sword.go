package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angelic Sell-Sword — Creature — Angel Mercenary {4}{W}, 4/4 (Outlaws of
// Thunder Junction Commander, #5368):
//
//	"Flying, vigilance
//	 Whenever this creature or another nontoken creature you control
//	 enters, create a 1/1 red Mercenary creature token with '{T}: Target
//	 creature you control gets +1/+0 until end of turn. Activate only as
//	 a sorcery.'
//	 Whenever this creature attacks, if its power is 6 or greater, draw a
//	 card."
//
// The enters trigger is one ability: the Sell-Sword itself, token or not,
// or any other nontoken creature entering under its controller. The attack
// trigger's "if its power is 6 or greater" is an intervening "if"
// (CR 603.4), read as the Sell-Sword is declared as an attacker and again
// as the ability resolves, from last-known information if it has left
// (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "60c77b18-c111-4e16-ad19-2ef32582cd9e",
		Name:            "Angelic Sell-Sword",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.Cmdr2S33ThisOrAnotherNontokenCreatureYouControlEntered,
				"Angelic Sell-Sword — create a 1/1 red Mercenary", effects.CreateAMercenary()),
			effects.InterveningIf(effects.Cmdr2S33CondThisPowerAtLeast{N: 6},
				effects.On(game.EventAttack, effects.ThisAttacked,
					"Angelic Sell-Sword — draw a card", effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}
