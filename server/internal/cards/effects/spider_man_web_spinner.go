package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Man, Web-Spinner — Legendary Creature — Spider Human Hero
// {3}{R}, 2/3 (Marvel Super Heroes Commander, #4455):
//
//	"Double strike
//	 Haste
//	 Whenever Spider-Man attacks, target creature can't block this turn."
//
// The target is any creature, chosen when the trigger goes on the stack;
// RestrictUntilEOT is the same restriction Ahn-Crop Crasher lays.
//
// No simplification.
func init() {
	const label = "Spider-Man, Web-Spinner — target creature can't block this turn"
	Register(Spec{
		OracleID:        "d586ef58-ca9e-4c6d-94d6-3b8d782b11e1",
		Name:            "Spider-Man, Web-Spinner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", "haste"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventAttack, ThisAttacked, label, targetCreatureCantBlockThisTurn(label)),
				TargetCreature("target creature")),
		},
	})
}
