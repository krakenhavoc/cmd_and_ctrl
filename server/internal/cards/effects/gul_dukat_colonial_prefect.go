package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gul Dukat, Colonial Prefect — Legendary Creature — Cardassian Rogue
// {4}{B}{B}{B}, 6/6 (Star Trek, #4973 batch 12):
//
//	"Menace
//	 At the beginning of each end step, each opponent loses life equal
//	 to the life that player lost this turn. You gain life equal to the
//	 life lost this way. (Damage causes loss of life.)"
//
// Archfiend of Despair's end-step sentence with a drain on the end. The
// amounts are read off the per-turn tally (b18LifeLostThisTurn) for
// every opponent before any is applied, so the loss this ability causes
// never inflates the next opponent's number.
//
// "The life lost this way" is what the opponents really lost, which
// differs from the sum of the amounts once something replaces a life
// loss. Each opponent's loss therefore goes through the keyed life-loss
// batch (LoseLifeEachKeyedForEffect), one opponent at a time because
// each has their own amount, and the running total rides a registered
// tail (gulDukatDrainNext) to the gain. A CR 616 ordering prompt over
// one opponent's loss is a restore point, and the gain waits for it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f41f0290-0f9d-4c51-935e-f69d2120e4a7",
		Name:            "Gul Dukat, Colonial Prefect",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, AnyPlayer,
				"Gul Dukat, Colonial Prefect — each opponent loses the life they lost this turn, you gain that much",
				gulDukatDrain),
		},
	})
}

func gulDukatDrain(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var players []uuid.UUID
	var amounts []int
	for _, opp := range ctx.Opponents() {
		if n := b18LifeLostThisTurn(g, opp); n > 0 {
			players = append(players, opp)
			amounts = append(amounts, n)
		}
	}
	return gulDukatDrainStep(g, game.KeyedTail{
		Ref: gulDukatDrainNext, Player: item.Controller, Source: ctx.Source(),
		Carry: players, Nums: amounts,
	})
}

// gulDukatDrainStep drains the first opponent left in t.Carry by the
// matching t.Nums amount, or, with none left, gains the running total
// t.N. Caller holds g.mu.
func gulDukatDrainStep(g *game.Game, t game.KeyedTail) error {
	if len(t.Carry) == 0 || len(t.Nums) == 0 {
		if t.N <= 0 {
			return nil
		}
		return g.ChangePlayerLifeForEffect(t.Source, t.Player, t.N)
	}
	next := t
	next.Carry = append([]uuid.UUID(nil), t.Carry[1:]...)
	next.Nums = append([]int(nil), t.Nums[1:]...)
	return g.LoseLifeEachKeyedForEffect(t.Source, []uuid.UUID{t.Carry[0]}, t.Nums[0], next)
}

// gulDukatDrainNext is told what one opponent really lost and moves on
// to the next. An on-disk key.
var gulDukatDrainNext = game.RegisterTail("gul-dukat/drain-each-opponent", func(g *game.Game, o game.TailOutcome) error {
	t := o.Tail
	if o.Amount > 0 {
		t.N += o.Amount
	}
	return gulDukatDrainStep(g, t)
})
