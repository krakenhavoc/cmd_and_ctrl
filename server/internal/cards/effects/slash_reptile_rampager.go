package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8abff7b5-fe27-46bd-b951-b5d9f93f3a14",
		Name:         "Slash, Reptile Rampager",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Slash — alliance: 2 damage to each opponent",
				Do(DamageEachOpponent{N: 2})),
			WheneverThisAttacks("Slash — create a 2/2 red Mutant creature token",
				Do(CreateToken{Template: TokenCard("2/2 red Mutant"), N: 1})),
		},
	})
}
