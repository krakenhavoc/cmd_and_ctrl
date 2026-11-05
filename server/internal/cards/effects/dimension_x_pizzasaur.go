package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dimension X Pizzasaur — {3}{B} Artifact Creature — Food Alien Mutant
// 2/1 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"When this creature enters, put two +1/+1 counters on target
//	 creature. When you do, destroy up to one target creature with mana
//	 value less than or equal to the number of counters among permanents
//	 you control.
//	 {2}, {T}, Sacrifice this creature: You gain 3 life and each opponent
//	 loses 3 life."
//
// The second sentence is a reflexive trigger (CR 603.12, the printed
// ruling): it is created only once counters were actually put on the
// target, and its own target is chosen as it goes on the stack. The
// bound is every counter of every kind on the permanents its controller
// controls, read live, so it is judged as the target is chosen and again
// as the reflexive trigger resolves (CR 608.2b). "Up to one" lets it
// resolve with no target. Food is an artifact type, never a creature
// type (CR 205.3g).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "29a9aff7-184c-459d-8002-e0287d978764",
		Name:         "Dimension X Pizzasaur",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Dimension X Pizzasaur — put two +1/+1 counters on target creature", pizzasaurCounters),
				TargetCreature("target creature")),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this creature: You gain 3 life and each opponent loses 3 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue}, // a drain answers nothing (ADR 0142)
			Cost:    game.AbilityCost{Mana: "{2}", Tap: true, SacrificeSelf: true},
			Effect:  Do(GainLife{Amount: 3}, EachOpponentLosesLife{N: 3}),
		}},
	})
}

const pizzasaurBefore = "pizzasaur-counters-before"

// pizzasaurCounters puts the counters, and only once they have landed
// (a CR 616 order may pause the placement) asks whether any did.
var pizzasaurCounters = Do(
	Instruction(func(ctx *Context) error {
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			c, _ := ctx.Game.LookupCardForEffect(t.ID)
			ctx.RecordStep(pizzasaurBefore, []uuid.UUID{t.ID}, c.Counters[game.CounterPlusOne], c.ObjectEpoch)
			return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
		}
		return nil
	}),
	Instruction(func(ctx *Context) error {
		rec := ctx.StepRecord(pizzasaurBefore)
		if len(rec.IDs) != 1 || len(rec.Ns) != 2 {
			return nil
		}
		c, ok := ctx.Game.LookupCardForEffect(rec.IDs[0])
		if !ok || !onBattlefield(ctx.Game, rec.IDs[0]) || c.ObjectEpoch != rec.Ns[1] || c.Counters[game.CounterPlusOne] <= rec.Ns[0] {
			return nil
		}
		return WhenYouDo("Dimension X Pizzasaur — destroy up to one target creature with mana value less than or equal to the number of counters among permanents you control",
			pizzasaurDestroyBody).Apply(ctx)
	}),
)

var pizzasaurDestroyBody = game.ReflexiveBody("dimension-x-pizzasaur/destroy", simpleBody(destroyFirstLegalCardTarget),
	constTargets(func() *game.TargetSpec {
		return TargetCreature("up to one target creature with mana value less than or equal to the number of counters among permanents you control",
			func(g *game.Game, you uuid.UUID, c game.Card) bool {
				mv, ok := g.ManaValueForEffect(c)
				return ok && mv <= countersOnPermanentsControlledBy(g, you)
			}).WithCount(0, 1)
	}))
