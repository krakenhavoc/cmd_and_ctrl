package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire Nation Engineer — Creature — Human Artificer {2}{B}, 2/3:
//
//	"Raid — At the beginning of your end step, if you attacked this
//	 turn, put a +1/+1 counter on another target creature or Vehicle you
//	 control."
//
// Raid is an ability word (CR 207.2c). The "if" is intervening (CR
// 603.4): the trigger is put on the stack, and its target asked, only
// when its controller declared an attacker this turn, and the condition
// is checked again as it resolves. Any attack counts, whoever it was at
// and whether or not the attacker is still around (the rulings). A
// Vehicle is a legal target crewed or not; "another" excludes the
// Engineer itself.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9a983666-376b-48a8-9f82-8d041d32ae93",
		Name:         "Fire Nation Engineer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, effects.TmcB6YouAttackedThisTurn),
				"Fire Nation Engineer — Raid: put a +1/+1 counter on another target creature or Vehicle you control",
				effects.Do(effects.If{Cond: effects.TmcB6CondYouAttacked{}, Then: []effects.Applier{
					effects.Instruction(func(ctx *effects.Context) error { return effects.PlusOneCounterOnTheTarget(ctx.Game, ctx.Item) }),
				}})),
				effects.Another(effects.TargetPermanent("another target creature or Vehicle you control",
					effects.Or(effects.Creature(), effects.Subtype("Vehicle")), effects.YouControl()))),
		},
	})
}
