package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agent 13, Sharon Carter — Legendary Creature — Human Spy Hero {2}{W}, 3/2
// (Marvel Super Heroes):
//
//	"Whenever a creature you control attacks alone, investigate."
//
// Exalted's trigger shape with a Clue for a payoff: the lone declaration is
// game.AttackedAlone, read when the attack trigger is harvested (after every
// attacker of the declaration is stamped), and the creature may be any you
// control, Agent 13 herself or not. A creature put onto the battlefield
// attacking was never declared, so it does not trigger this.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ddf7201a-6672-416d-be6b-202dafbeccab",
		Name:         "Agent 13, Sharon Carter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, mshB1YouAttackAlone, "Agent 13, Sharon Carter — investigate",
				Do(CreateToken{Template: ClueToken(), N: 1})),
		},
	})
}
