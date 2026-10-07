package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rangers of Ithilien — Creature — Human Ranger {2}{U}{U}, 3/3:
//
//	"Vigilance
//	 When this creature enters, gain control of up to one target
//	 creature with lesser power for as long as you control this
//	 creature. Then the Ring tempts you."
//
// The clause is relative to the Rangers (#2146,
// RelativeToSource(LesserPower())): its power is compared with
// the target's as the target is chosen and again as the ability
// resolves (2023-06-16 ruling), so a Rangers shrunk in response, or one
// that has left the battlefield (its last-known power, CR 608.2h),
// leaves the target illegal. "Up to one" lets the trigger go on the
// stack with no target; the Ring still tempts you either way. The
// theft lasts while you control the Rangers (CR 611.2b) and never
// begins if it has already left.
//
// No simplification.
func init() {
	const label = "Rangers of Ithilien — gain control of up to one target creature with lesser power"
	steal := GainControlOfTargetFor(label, func(ctx *Context) (game.Duration, bool) {
		return DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller())
	})
	Register(Spec{
		OracleID:        "4d5fd9f0-ecb3-4b91-ae55-f02336d7cf37",
		Name:            "Rangers of Ithilien",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters(label+". Then the Ring tempts you", func(g *game.Game, item *game.StackItem) error {
					if err := steal(g, item); err != nil {
						return err
					}
					return TheRingTemptsYou{}.Apply(NewContext(g, item))
				}),
				RelativeToSource(
					TargetCreature("up to one target creature with lesser power").WithCount(0, 1),
					LesserPower()),
			),
		},
	})
}
