package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kinetic Ooze — Creature — Ooze {X}{G}, 0/0:
//
//	"This creature enters with X +1/+1 counters on it.
//	 When this creature enters, destroy up to one target artifact or
//	 enchantment with mana value X or less. If X is 5 or more, you draw
//	 a card. If X is 10 or more, double the number of +1/+1 counters on
//	 any number of other target creatures."
//
// X is the value announced for the spell (CR 107.3m), read off the
// permanent as the trigger is built and carried on the item, so it is
// exact even if the Ooze has left by the time the trigger resolves; an
// Ooze that was not cast has X = 0. The trigger's clauses are built from
// that X: the destroy clause is bounded by it, and the second clause,
// "any number of other target creatures", exists only when X is 10 or
// more. Doubling gives each still-legal creature as many +1/+1 counters
// as it has as the trigger resolves (CR 701.10e).
//
// No simplification.
func init() {
	const label = "Kinetic Ooze — destroy up to one target artifact or enchantment with mana value X or less; draw if X is 5 or more; double counters if X is 10 or more"
	trig := game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventETB},
		AppliesTo: Self,
		Key:       label,
		TargetsFrom: func(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
			x := source.CastX()
			destroy := TargetPermanent("up to one target artifact or enchantment with mana value X or less",
				Or(Artifact(), Enchantment()), ManaValueLE(x)).WithCount(0, 1)
			if x < 10 {
				return destroy
			}
			return Clauses(destroy, Another(TargetCreature("any number of other target creatures").WithCount(0, 0)))
		},
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Amount = source.CastX()
			return item
		},
		Effect: kineticOozeResolve,
	}
	Register(Spec{
		OracleID:                   "ab74c6b1-2099-4429-94b7-a5656ac04db3",
		Name:                       "Kinetic Ooze",
		Completeness:               CompletenessFull,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		XMatters:                   true,
		Triggered:                  []game.TriggeredAbility{trig},
	})
}

// kineticOozeResolve is the three sentences as steps of one walk, so a
// later sentence waits for any prompt the destruction's replacement
// window asks (CR 608.2c); none of them depends on the destruction.
var kineticOozeResolve = Do(
	Instruction(kineticOozeDestroy),
	Instruction(kineticOozeDraw),
	Instruction(kineticOozeDouble),
)

// kineticOozeDestroy destroys the first clause's pick, if it is still legal.
func kineticOozeDestroy(ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return DestroyTarget{Target: t.ID}.Apply(ctx)
}

// kineticOozeDraw is "If X is 5 or more, you draw a card."
func kineticOozeDraw(ctx *Context) error {
	if ctx.Item.Params.Amount < 5 {
		return nil
	}
	return DrawCards{N: 1}.Apply(ctx)
}

// kineticOozeDouble is "If X is 10 or more, double the number of +1/+1
// counters on any number of other target creatures" (CR 701.10e).
func kineticOozeDouble(ctx *Context) error {
	if ctx.Item.Params.Amount < 10 {
		return nil
	}
	for _, t := range ctx.ClauseTargets(1) {
		if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			continue
		}
		if n := c.Counters[game.CounterPlusOne]; n > 0 {
			if err := nothingIfGone(AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)); err != nil {
				return err
			}
		}
	}
	return nil
}
