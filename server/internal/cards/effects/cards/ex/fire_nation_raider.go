package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire Nation Raider — Creature — Human Soldier {3}{R}, 4/2 (Avatar: The
// Last Airbender, #4459):
//
//	"Raid — When this creature enters, if you attacked this turn, create
//	 a Clue token."
//
// Raid is an intervening "if" (CR 603.4): checked as it triggers and
// again as it resolves. Any attack declared this turn counts, whatever
// has happened to the attacker since (the rulings).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b37e9bc5-6a2c-4f08-b856-3f8d0d16defb",
		Name:         "Fire Nation Raider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.TmcB6YouAttackedThisTurn),
				"Fire Nation Raider — Raid: create a Clue token",
				effects.Do(effects.If{Cond: effects.AvatarB11CondYouAttacked{}, Then: []effects.Applier{
					effects.CreateToken{Template: effects.ClueToken(), N: 1},
				}})),
		},
	})
}
