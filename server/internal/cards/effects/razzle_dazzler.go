package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Razzle-Dazzler — Creature — Human Wizard {1}{U}, 1/2 (Outlaws of
// Thunder Junction, #4468):
//
//	"Whenever you cast your second spell each turn, put a +1/+1 counter
//	 on this creature. It can't be blocked this turn."
//
// The trigger is Devoted Duelist's YouCastYourSecondSpellEachTurn (the
// per-turn cast tally reads two), and the two riders are Kappa
// Cannoneer's: the counter on this object, then RestrictUntilEOT's
// can't-be-blocked, pinned to it until the turn ends. A Dazzler that
// left and came back is a new object (CR 400.7) and gets neither.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a1ac6d5c-b2d0-4ca4-8af8-16fd3dd3b3b5",
		Name:         "Razzle-Dazzler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn, "Razzle-Dazzler — +1/+1 counter; it can't be blocked this turn", Do(
				// The restriction waits for a prompt the counter asks (CR 608.2c, #4155).
				Instruction(func(ctx *Context) error {
					return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}),
				Instruction(func(ctx *Context) error {
					return RestrictUntilEOT{
						Target:       ctx.Source(),
						Restrictions: game.CantBeBlocked,
						Label:        "Razzle-Dazzler — can't be blocked",
					}.Apply(ctx)
				}),
			)),
		},
	})
}
