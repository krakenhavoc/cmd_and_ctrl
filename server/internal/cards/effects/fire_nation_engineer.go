package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9a983666-376b-48a8-9f82-8d041d32ae93",
		Name:         "Fire Nation Engineer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventBeginEndStep, AllOf(ByYou, tmcB6YouAttackedThisTurn),
				"Fire Nation Engineer — Raid: put a +1/+1 counter on another target creature or Vehicle you control",
				Do(If{Cond: tmcB6CondYouAttacked{}, Then: []Applier{
					Instruction(func(ctx *Context) error { return PlusOneCounterOnTheTarget(ctx.Game, ctx.Item) }),
				}})),
				Another(TargetPermanent("another target creature or Vehicle you control",
					Or(Creature(), Subtype("Vehicle")), YouControl()))),
		},
	})
}
