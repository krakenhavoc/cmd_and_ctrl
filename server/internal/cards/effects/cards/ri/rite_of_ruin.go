package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rite of Ruin — {5}{R}{R} Sorcery: "Choose an order for artifacts, creatures,
// and lands. Each player sacrifices one permanent of their choice of the first
// type, sacrifices two of their choice of the second type, then sacrifices three
// of their choice of the third type."
//
// The caster picks one of the six orders. Each step has every player sacrifice
// its number of that type together, and the next step starts only after those
// are gone, so an artifact creature sacrificed for the artifacts is no longer
// there for the creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0dd0586-d89e-40e9-863d-9c390bee117a",
		Name:         "Rite of Ruin",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return riteOfRuinChooseOrder(ctx.Game, item)
		},
	})
}

var riteOfRuinTypes = []struct {
	label string
	pred  effects.CardPredicate
}{
	{"artifacts", effects.Artifact()},
	{"creatures", effects.Creature()},
	{"lands", effects.Land()},
}

// riteOfRuinOrders are the six orders of artifacts, creatures and lands, as
// indexes into riteOfRuinTypes.
var riteOfRuinOrders = [6][3]int{
	{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0},
}

// riteOfRuinChooseOrder is Rite of Ruin's first sentence: its controller
// chooses an order for artifacts, creatures and lands.
func riteOfRuinChooseOrder(g *game.Game, item *game.StackItem) error {
	opts := make([]game.ChoiceOption, len(riteOfRuinOrders))
	for i, o := range riteOfRuinOrders {
		opts[i] = game.ChoiceOption{Label: "1. " + riteOfRuinTypes[o[0]].label + ", 2. " + riteOfRuinTypes[o[1]].label + ", 3. " + riteOfRuinTypes[o[2]].label}
	}
	return effects.PickOption{
		Question: "Rite of Ruin — choose an order for artifacts, creatures, and lands",
		Options:  opts,
		ThenKey:  riteOfRuinOrderChosen,
		WithItem: true,
	}.Apply(effects.NewContext(g, item))
}

var riteOfRuinOrderChosen = effects.OptionPickThen("option-pick/s82-old-b435-rite-of-ruin-order", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(riteOfRuinOrders) {
		return nil
	}
	return riteOfRuinStep(ctx, r.Index, 0)
})

// riteOfRuinStep has each player sacrifice step+1 permanents of the step's
// type (one of the first, two of the second, three of the third), then moves
// on to the next type once everything chosen has been sacrificed.
func riteOfRuinStep(ctx *effects.Context, order, step int) error {
	if step >= 3 {
		return nil
	}
	t := riteOfRuinTypes[riteOfRuinOrders[order][step]]
	var asks []uuid.UUID
	for _, p := range ctx.Game.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		for i := 0; i <= step; i++ {
			asks = append(asks, p.ID)
		}
	}
	return ctx.PlayersSacrificeThen(asks, effects.SacrificeSpec("an "+t.label[:len(t.label)-1], t.pred),
		"Rite of Ruin — sacrifice "+riteOfRuinCount(step)+" of your "+t.label, riteOfRuinNext, ctx.RunData(nil, order, step))
}

func riteOfRuinCount(step int) string {
	return []string{"one", "two", "three"}[step]
}

var riteOfRuinNext game.PromptRunThen

func init() {
	riteOfRuinNext = effects.RunThen("prompt-run/s82-old-b435-rite-of-ruin-next", func(ctx *effects.Context, r game.RunDone) error {
		if len(r.Nums) < 2 {
			return nil
		}
		return riteOfRuinStep(ctx, r.Nums[0], r.Nums[1]+1)
	})
}
