package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Force Bubble — Enchantment {2}{W}{W}:
//
//	"If damage would be dealt to you, put that many depletion counters
//	 on this enchantment instead.
//	 When there are four or more depletion counters on this enchantment,
//	 sacrifice it.
//	 At the beginning of each end step, remove all depletion counters
//	 from this enchantment."
//
// ADR 0107 §1 (#1858). The first line is a CR 614.1a replacement ("instead"),
// not a prevention effect: it does not say "prevent" (CR 615.1a). Damage
// to the controller, combat or not, becomes that many counters. The
// sacrifice is a CR 603.8 state trigger: a single 5-damage hit puts five
// counters on and triggers it once, and the Bubble still absorbs whatever
// arrives before the trigger resolves. Every end step — anyone's —
// empties it again.
//
// No simplification.
func init() {
	const depletion = "depletion"
	Register(Spec{
		OracleID:     "655ae8e7-372b-4d8d-b33f-4aca46831abb",
		Name:         "Force Bubble",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			forceBubbleDepletion(),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisHasAtLeast(depletion, 4, "Force Bubble — sacrifice it", SacrificeThisIfStillOnBattlefield),
			AtEachStep(game.StepEnd, "Force Bubble — remove all depletion counters", removeAllCountersFromThis(depletion)),
		},
	})
}

// forceBubbleDepletion is "If damage would be dealt to you, put that many
// depletion counters on this enchantment instead": damage to the
// Bubble's controller is cancelled and that many counters go on it. The
// event is the damage to the PLAYER (CR 616.1: the affected player, the
// controller here, orders it against other replacements), combat and
// noncombat alike. A Bubble the counters cannot land on still cancels
// the damage: the replacement applied, and CR 614.6 says the replaced
// event never happens.
//
// Prevention is declared FALSE (ADR 0107 §5): the text says "instead",
// not "prevent" (CR 614.1a, 615.1a), so damage that can't be prevented
// (CR 615.12) is still turned into depletion counters.
func forceBubbleDepletion() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return src != nil && ev.Kind == game.RepEventDamage && ev.DamageAmount > 0 &&
				ev.DamageTarget == src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			n := ev.DamageAmount
			ev.Cancel()
			return g.AddCounterForEffect(src.InstanceID, "depletion", n)
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Prevention: false,
		Label:      "Force Bubble — put that many depletion counters on it instead",
	}
}
