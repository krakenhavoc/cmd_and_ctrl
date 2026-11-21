package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emptiness — Creature — Elemental Incarnation {4}{W/B}{W/B}, 3/5
// (Lorwyn Eclipsed, #4458):
//
//	"When this creature enters, if {W}{W} was spent to cast it, return
//	 target creature card with mana value 3 or less from your graveyard
//	 to the battlefield.
//	 When this creature enters, if {B}{B} was spent to cast it, put
//	 three -1/-1 counters on up to one target creature.
//	 Evoke {W/B}{W/B}"
//
// Deceit's cycle-mate: two enters triggers behind intervening ifs
// (CR 603.4) on the mana spent (deceitSpent). The return is Reaping
// Willow's: a target in its controller's own graveyard, chosen as the
// trigger is put on the stack (CR 603.3d), and an illegal target by
// resolution returns nothing. The counters go on as an effect, through
// the CR 614 counter window, so Doubling Season doubles them. "Up to
// one" lets the trigger target nothing.
//
// Evoked for {W}{W}, Emptiness can return itself only if it is already
// in the graveyard when the trigger is put on the stack, which it isn't:
// the evoke sacrifice is its own trigger, ordered with this one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "dc5dd7f7-375e-4d51-85cb-1abf16ca3854",
		Name:           "Emptiness",
		ReadsManaSpent: game.ReadsOwnPayment,
		Completeness:   effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Evoke("{W/B}{W/B}"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.DeceitSpent("W")),
					"Emptiness — return target creature card with mana value 3 or less from your graveyard to the battlefield",
					effects.ReturnFirstLegalGraveyardTargetToBattlefield),
				effects.TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard", effects.YouOwn(), effects.Creature(), effects.ManaValueLE(3))),
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.DeceitSpent("B")),
					"Emptiness — put three -1/-1 counters on up to one target creature",
					emptinessCounters),
				effects.TargetCreature("up to one target creature").WithCount(0, 1)),
		},
	})
}

// emptinessCounters puts three -1/-1 counters on the target, if any.
func emptinessCounters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return effects.AddCounter{Target: t.ID, Kind: game.CounterMinusOne, N: 3}.Apply(ctx)
		}
	}
	return nil
}
