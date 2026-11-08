package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Searslicer Goblin — Creature — Goblin Warrior {1}{R}, 2/1
// (Foundations, #4465):
//
//	"Raid — At the beginning of your end step, if you attacked this
//	 turn, create a 1/1 red Goblin creature token."
//
// Fire Nation Engineer's raid end-step shape with Dragon Fodder's
// Goblin. Raid is an intervening "if" (CR 603.4): the trigger goes on
// the stack only when its controller declared an attacker this turn,
// and the condition is checked again as it resolves. Any attack counts,
// whatever has happened to the attacker since.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "86ec0b56-497f-481c-a03a-c4e64e8e7404",
		Name:         "Searslicer Goblin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, AllOf(ByYou, tmcB6YouAttackedThisTurn),
				"Searslicer Goblin — Raid: create a 1/1 red Goblin creature token",
				Do(If{Cond: avatarB11CondYouAttacked{}, Then: []Applier{
					CreateToken{Template: RedGoblinToken(), N: 1},
				}})),
		},
	})
}
