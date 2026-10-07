package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vodalian Mindsinger — Creature — Merfolk Wizard {1}{U}{U}, 2/2:
//
//	"Kicker {1}{R} and/or {1}{G}
//	 This creature enters with two +1/+1 counters on it for each time
//	 it was kicked.
//	 When this creature enters, gain control of target creature with
//	 power less than this creature's power for as long as you control
//	 this creature."
//
// Two kicker costs (CR 702.33b, #2153) that share one clause, so the
// counters count both together (CR 702.33d): CountersPerKick on the
// CR 614 entry pipeline, which Doubling Season doubles.
//
// "Power less than this creature's power" is a clause relative to the
// Mindsinger (#2146, RelativeToSource), judged against its power as the
// trigger goes on the stack and again at resolution (CR 608.2b), so a
// Mindsinger pumped or shrunk in response moves the bound. The counters
// are already on it by then, so a
// kicked Mindsinger reaches bigger creatures, as printed. The theft
// lasts while the controller still controls the Mindsinger
// (CR 611.2b): it ends if the Mindsinger leaves or changes hands, and
// never begins if the Mindsinger has already left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "3823afcf-6b15-46a9-ba49-3a9a39f8114e",
		Name:                       "Vodalian Mindsinger",
		Completeness:               CompletenessFull,
		OptionalCosts:              Kickers("{1}{R}", "{1}{G}"),
		EntersWithCountersFromCast: []game.EntryCountersFromCast{CountersPerKick(game.CounterPlusOne, 2)},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       "Vodalian Mindsinger — gain control of target creature with lesser power",
			Targets: RelativeToSource(
				TargetCreature("target creature with power less than this creature's power"),
				LesserPower()),
			Effect: GainControlOfTargetFor("Vodalian Mindsinger — gain control while you control it", func(ctx *Context) (game.Duration, bool) {
				return DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller())
			}),
		}},
	})
}
