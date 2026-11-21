package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stature, Size Shifter — {U} Legendary Creature — Human Hero, 1/1 (Marvel
// Super Heroes, #4455):
//
//	"Stature can't be blocked if her power is 1 or less.
//	 Power-up — {X}{U}{U}: Put X +1/+1 counters on Stature."
//
// The evasion is Thieves' Tools' conditional "can't be blocked" block rule,
// read off her live power when blockers are declared (CR 509.1b), after
// every layer has run. Power-up is the engine's (CR 702.193): the turn she
// enters it costs {X}{U}{U} less {U}; X is announced with the activation and
// read back with ctx.X().
//
// No simplification.
func init() {
	powerUp := effects.PowerUp("{X}{U}{U}: Put X +1/+1 counters on Stature.", effects.ManaCost("{X}{U}{U}"),
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			return effects.CounterOnThis{Kind: game.CounterPlusOne, N: ctx.X()}.Apply(ctx)
		})
	powerUp.Purpose.Answers = game.AnswerPump
	effects.Register(effects.Spec{
		OracleID:     "bd49a05c-ed51-4d69-ab9b-11192442ec17",
		Name:         "Stature, Size Shifter",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedWhile(effects.OnSelf(), effects.PowerLE(1)),
		},
		Activated: []effects.ActivatedAbility{powerUp},
	})
}
