package cardsn

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Numa, Joraga Chieftain — Legendary Creature — Elf Warrior {2}{G}, 2/2
// (Kaldheim Commander, #5506):
//
//	"At the beginning of combat on your turn, you may pay {X}{X}. When you
//	 do, distribute X +1/+1 counters among any number of target Elves.
//	 Partner"
//
// X is chosen as the ability resolves (MayPayX over a cost with two {X},
// so paying X buys {X}{X}), and the payment is the optional cost. "When you
// do" is a CR 603.12 reflexive trigger: its targets (any number of Elves,
// at most X of them) and the split are announced as it goes on the stack
// (CR 601.2d), each target getting at least one counter. Nothing is created
// for X = 0. Partner is a deck-construction rule.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f4aa98d-3572-4637-9b0d-3c12f63a96ec",
		Name:         "Numa, Joraga Chieftain",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Numa, Joraga Chieftain — you may pay {X}{X}; when you do, distribute X +1/+1 counters among any number of target Elves",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPayX{
						Cost:  "{X}{X}",
						Label: "Numa, Joraga Chieftain",
						Buys:  "distribute X +1/+1 counters among target Elves",
						Then:  numaPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// numaPaid is the paid branch: the reflexive trigger carrying X. An on-disk
// key.
var numaPaid = effects.PayUnlessBranches("pay-unless/numa-joraga-chieftain-paid",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return effects.ReflexiveTrigger{
			Label:  "Numa, Joraga Chieftain — distribute " + strconv.Itoa(r.Number) + " +1/+1 counters among any number of target Elves",
			Body:   numaDistributeBody,
			Params: game.EffectParams{Amount: r.Number},
		}.Apply(ctx)
	}, nil)

// numaDistributeBody is the reflexive trigger: the clause is any number of
// target Elves (up to X, one counter each at least) with X divided among
// them. An on-disk key.
var numaDistributeBody = game.ReflexiveBody("numa-joraga-chieftain/distribute",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		return effects.PutDividedCounters(effects.NewContext(g, item), game.CounterPlusOne)
	}),
	func(sourceID uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return effects.TargetPermanent("any number of target Elves", effects.Creature(), effects.HasSubtype("Elf")).WithCount(0, p.Amount).Dividing(effects.Divide(p.Amount))
	})
