package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mistford River Turtle — Creature — Turtle {3}{U}, 1/5 (Throne of Eldraine,
// S82 old sets):
//
//	"Whenever this creature attacks, another target attacking non-Human
//	 creature can't be blocked this turn."
//
// Passenger Ferry's grant with a non-Human clause: the target is chosen as the
// trigger goes on the stack, "another" excludes the Turtle, "attacking" and
// "non-Human" are read when it is chosen and again as it resolves, and a
// target that is gone is left alone (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Mistford River Turtle — another target attacking non-Human creature can't be blocked this turn"
	attack := effects.WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		target := effects.FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		return effects.RestrictUntilEOT{Target: target, Restrictions: game.CantBeBlocked, Label: label}.Apply(ctx)
	})
	attack.Targets = effects.Another(effects.TargetCreature("another target attacking non-Human creature", effects.AttackingCreature(), effects.Not(effects.Subtype("Human"))))
	effects.Register(effects.Spec{
		OracleID:     "c5e9389e-6ee3-48df-89bc-9334544e0a4c",
		Name:         "Mistford River Turtle",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{attack},
	})
}
