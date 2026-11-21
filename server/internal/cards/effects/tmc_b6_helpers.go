package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmc_b6_helpers.go — shared pieces of TMNT Commander batch 6 (#5011).

// tmcB6DoubleDamageFromCreaturesWithCounters is Raphael, the Muscle's
// "Double all damage that creatures you control with counters on them
// would deal": a damage replacement (CR 701.10g) on the SOURCE, a creature
// the replacement's controller controls that has at least one counter of
// any kind on it as the damage would be dealt, combat and noncombat
// alike, to any permanent or player. The source's type and controller are
// read from the event's last-known characteristics (CR 608.2h,
// damageSourceCharacteristics), and its counters from the battlefield
// object, or, for a creature that has already left, from the counters it
// last had there. The amount is divided or assigned before it is doubled
// (the card's ruling): the event already carries each recipient's share.
// With a second doubler the affected player orders them (CR 616.1).
func TmcB6DoubleDamageFromCreaturesWithCounters(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil {
				return false
			}
			ch, ok := DamageSourceCharacteristics(ev, g)
			if !ok || ch.Controller != src.Controller || !HasFold(ch.Types, "Creature") {
				return false
			}
			return tmcB6HasACounter(g, ev.DamageSource)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.DamageAmount *= 2
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// tmcB6HasACounter reports whether the damage source has one or more
// counters of any kind on it: the permanent's own while it is on the
// battlefield, its last-known counters once it has left.
func tmcB6HasACounter(g *game.Game, id uuid.UUID) bool {
	counters := g.LastKnownCountersForEffect(id)
	if OnBattlefield(g, id) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return false
		}
		counters = c.Counters
	}
	for _, n := range counters {
		if n > 0 {
			return true
		}
	}
	return false
}

// tmcB6YouAttackedThisTurn is the trigger-time half of "if you attacked
// this turn" (raid): the source's controller declared at least one
// attacker this turn.
func TmcB6YouAttackedThisTurn(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return B18AttackedThisTurn(g, source.Controller)
}

// tmcB6CondYouAttacked is the resolution-time half (CR 603.4).
type TmcB6CondYouAttacked struct{}

func (TmcB6CondYouAttacked) Holds(ctx *Context) bool {
	return B18AttackedThisTurn(ctx.Game, ctx.Controller())
}

// condition declares no footprint: the attack tally has no
// FootprintCond, and the row it guards has a target, so it would not
// have one anyway.
func (TmcB6CondYouAttacked) Condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}
