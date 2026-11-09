package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Broodguard Elite — Creature — Insect Knight {X}{G}{G}, 0/0 (Edge of
// Eternities, #4461):
//
//	"This creature enters with X +1/+1 counters on it.
//	 When this creature leaves the battlefield, put its counters on
//	 target creature you control.
//	 Warp {X}{G} (You may cast this card from your hand for its warp
//	 cost. Exile this creature at the beginning of the next end step,
//	 then you may cast it from exile on a later turn.)"
//
// The entry counters are Feral Hydra's XCounters: the X announced for
// the spell, whichever cost paid it, so a warped Elite enters with as
// many counters as one cast for its printed cost with the same X.
//
// The leaves trigger fires on every exit (dying, exile, bounce), the
// warp end-step exile included, and reads "its counters" from the
// Elite's last-known information (ctx.TriggeringPermanent, CR 603.10),
// every kind, as Sin, Unending Cataclysm does. So the warp line is a
// one-turn body that hands its counters to another creature at the end
// step. The Elite is gone by the time the target is chosen, so it can
// never be its own target; with no creature to target the trigger is
// removed, and a pick that is no longer legal gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "c90da077-de8c-4402-abfc-b63dd15853da",
		Name:                       "Broodguard Elite",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		AlternativeCosts:           []game.AlternativeCost{Warp("{X}{G}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisLeaves("Broodguard Elite — put its counters on target creature you control",
				stdB9PutLeftCountersOnTarget),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
