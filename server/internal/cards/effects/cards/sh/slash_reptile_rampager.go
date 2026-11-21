package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slash, Reptile Rampager — {3}{R}{R} Legendary Creature — Mutant Berserker
// Turtle 7/5 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Alliance — Whenever another creature you control enters, Slash deals
//	 2 damage to each opponent.
//	 Whenever Slash attacks, create a 2/2 red Mutant creature token."
//
// Alliance is an ability word (CR 207.2c): the trigger is "whenever another
// creature you control enters", one per creature, tokens included. The
// damage is one event from Slash to every opponent still in the game. The
// attack trigger's Mutant is not attacking, and its arrival triggers the
// alliance ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8abff7b5-fe27-46bd-b951-b5d9f93f3a14",
		Name:         "Slash, Reptile Rampager",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverAnotherCreatureEntersUnderYourControl("Slash — alliance: 2 damage to each opponent",
				effects.Do(effects.DamageEachOpponent{N: 2})),
			effects.WheneverThisAttacks("Slash — create a 2/2 red Mutant creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 red Mutant"), N: 1})),
		},
	})
}
