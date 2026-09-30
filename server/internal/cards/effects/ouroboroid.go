package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ouroboroid — Creature — Plant Wurm {2}{G}{G}, 1/3 (slice 296-m):
//
//	"At the beginning of combat on your turn, put X +1/+1 counters on
//	 each creature you control, where X is this creature's power."
//
// AtBeginningOfYourCombat (Goblin Rabblemaster's own step trigger)
// with an effect that reads Ouroboroid's CURRENT power off its
// effective characteristics — counters and anthems already on it
// count, exactly as the printed card means "this creature's power" at
// the moment the trigger resolves, not its base 1.
//
// If Ouroboroid itself has left the battlefield by resolution, X is
// read from nothing (the source is gone) and the trigger does
// nothing, which is the CR 608.2b outcome for an ability whose whole
// effect depends on a source that is no longer there to measure.
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
	source, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	x := source.Effective().Power
	if x <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
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
