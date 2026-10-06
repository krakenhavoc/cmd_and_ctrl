package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ouroboroid — Creature — Plant Wurm {2}{G}{G}, 1/3 (slice 296-m):
//
//	"At the beginning of combat on your turn, put X +1/+1 counters on
//	 each creature you control, where X is this creature's power."
//
// AtBeginningOfYourCombat (Goblin Rabblemaster's own step trigger)
// with an effect that reads Ouroboroid's power once, as the trigger
// resolves (CR 608.2h), and puts that many counters on each creature.
//
// "This creature's power" is the whole power: the +1/+1 counters on
// it count (CR 122.1a, applied in layer 7c, CR 613.4c). #2401: the
// first version read Effective().Power, which stops before the
// counters, so X stayed at the printed 1 forever and the card never
// grew. ctx.SourcePermanent's Power is counter-aware.
//
// If Ouroboroid has left the battlefield by then, X is its power as it
// last existed there, counters included (CR 608.2h's last-known
// information), and the other creatures still get their counters. A
// negative power makes X zero (CR 107.1b) and puts nothing.
//
// X is fixed before the first counter goes on, so Ouroboroid's own
// counters from this resolution do not change what the rest get, and a
// replacement such as Hardened Scales applies to each placement
// separately (each creature's counters are their own event).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50d6fd91-23d3-4d32-804f-6233e4386904",
		Name:         "Ouroboroid",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Ouroboroid — put counters on each creature you control equal to its power",
				ouroboroidCombatCounters),
		},
	})
}

func ouroboroidCombatCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x, ok := ouroboroidPower(g, ctx)
	if !ok {
		return nil
	}
	if x <= 0 {
		return nil
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller || !c.IsCreature() {
			continue
		}
		if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: x}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// ouroboroidPower is "this creature's power" at resolution: the source
// object's live power, or its last-known power once it has left. An
// item with no source object stamped (a restore point written before
// StackItem.SourceObject existed) falls back to the card on the
// battlefield, if it is there.
func ouroboroidPower(g *game.Game, ctx *Context) (int, bool) {
	if info, ok := ctx.SourcePermanent(); ok {
		return info.Power, true
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == ctx.Source() {
			return c.PowerForComparison(), true
		}
	}
	return 0, false
}
