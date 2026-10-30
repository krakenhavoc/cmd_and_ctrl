package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Batroc the Leaper — Legendary Creature — Human Villain {1}{R}, 2/2
// (Marvel Super Heroes Commander, #4455):
//
//	"Multikicker {2}
//	 Batroc enters with a +1/+1 counter on him for each time he was
//	 kicked.
//	 When Batroc enters, he deals damage equal to his power to each of up
//	 to X targets, where X is the number of times he was kicked."
//
// Multikicker and the entry counters are Everflowing Chalice's. The
// trigger's target group is built as it goes on the stack from the
// source's cast provenance, the way Blitzball Stadium's "up to X" is
// (a Batroc that was not cast was kicked no times and asks for nothing).
// Damage is his power as the ability resolves (CR 608.2h: last-known
// information if he has left), counters included, dealt as one instance.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Batroc the Leaper — damage equal to his power to each of up to X targets",
		sourceDealsComputedDamageToEachLegalTarget(func(ctx *Context) int {
			info, ok := ctx.SourcePermanent()
			if !ok || info.Power < 0 {
				return 0
			}
			return info.Power
		}))
	enters.TargetsFrom = func(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
		x := game.CardKickedTimes(*source)
		if x <= 0 {
			return nil
		}
		return TargetAny().WithCount(0, x)
	}
	Register(Spec{
		OracleID:     "ebbea7a7-c990-43fa-8f67-a5da17fbbb10",
		Name:         "Batroc the Leaper",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{
			Multikicker("{2}", 20),
		},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			CountersPerKick(game.CounterPlusOne, 1),
		},
		Triggered: []game.TriggeredAbility{enters},
	})
}
