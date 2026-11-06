package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iron-Fist Pulverizer — Creature — Giant Warrior {4}{R}, 4/5 (Outlaws
// of Thunder Junction, #4468):
//
//	"Reach
//	 Whenever you cast your second spell each turn, this creature deals
//	 2 damage to target opponent. Scry 1. (Look at the top card of your
//	 library. You may put that card on the bottom.)"
//
// The trigger is Alphinaud Leveilleur's YouCastYourSecondSpellEachTurn.
// The opponent is chosen as the trigger goes on the stack (CR 603.3d);
// the damage is dealt by the Pulverizer and the scry follows it. An
// opponent who has left the game by resolution makes the ability's only
// target illegal, so it does nothing at all, scry included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e5cea282-9e61-47da-8c5d-f5fca6a508a4",
		Name:            "Iron-Fist Pulverizer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Iron-Fist Pulverizer — 2 damage to target opponent, then scry 1",
				ironFistPulverizerDamageThenScry),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// ironFistPulverizerDamageThenScry is the trigger's resolution: 2 damage
// from the Pulverizer to the target opponent, then scry 1.
func ironFistPulverizerDamageThenScry(g *game.Game, item *game.StackItem) error {
	if err := DealDamageToTheTarget(2)(g, item); err != nil {
		return err
	}
	return Scry{N: 1}.Apply(NewContext(g, item))
}
