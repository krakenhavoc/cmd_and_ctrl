package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grave Endeavor — Instant {5}{B}{B} (Adventures in the Forgotten Realms):
//
//	"Roll two d10 and choose one result. Return a creature card from your
//	 graveyard to the battlefield with a number of +1/+1 counters on it
//	 equal to that result. Then each opponent loses X life and you gain X
//	 life, where X is the other result."
//
// The shared two-dice shape is afrEndeavorTail. The creature is chosen,
// not targeted; the counters are an "enters with" clause (CR 614.1c) and
// the drain is the return's own tail, so it follows the entry. With no
// creature card in the graveyard the drain still happens.
func init() {
	Register(Spec{
		OracleID: "917dc5a0-071d-4658-ab58-29ac841260f7", Name: "Grave Endeavor", Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return RollDice{Sides: 10, N: 2, Then: graveEndeavorRolled}.Apply(ctx)
		},
	})
}

var graveEndeavorRolled = afrEndeavorTail("dice/grave-endeavor", "Grave Endeavor",
	func(chosen, other int) string {
		return strconv.Itoa(chosen) + " counters, drain " + strconv.Itoa(other)
	},
	func(ctx *Context, chosen, other int) error {
		g := ctx.Game
		candidates := graveyardIDs(g, ctx.Controller(), func(c game.Card) bool { return c.IsCreature() })
		if len(candidates) == 0 {
			return graveEndeavorDrainLife(ctx, other)
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:   ctx.Controller(),
			Source:    ctx.Source(),
			Question:  "Grave Endeavor — return a creature card from your graveyard to the battlefield with " + strconv.Itoa(chosen) + " +1/+1 counters",
			Cards:     candidates,
			Min:       1,
			Max:       1,
			Zone:      game.ZoneGraveyard,
			ThenKey:   graveEndeavorPicked,
			Asked:     game.NewAskedItem(ctx.Item),
			LiveZone:  game.ZoneGraveyard,
			CarryNums: []int{chosen, other},
		})
		return nil
	})

var graveEndeavorPicked = askedPick("choose-cards/grave-endeavor-pick", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Nums) < 2 {
		return nil
	}
	chosen, other := r.Nums[0], r.Nums[1]
	if len(r.Picked) == 0 {
		return graveEndeavorDrainLife(ctx, other)
	}
	tail := withN(ctx.ItemTail(graveEndeavorDrain), other)
	return ctx.Game.ReturnFromGraveyardWithCountersKeyedForEffect(r.Picked[0], uuid.Nil, false,
		map[string]int{game.CounterPlusOne: chosen}, tail)
})

// graveEndeavorDrain is "then each opponent loses X life and you gain X
// life" once the creature has entered: X is the tail's N.
var graveEndeavorDrain = CardTail("then/grave-endeavor-drain", func(ctx *Context, o game.TailOutcome) error {
	return graveEndeavorDrainLife(ctx, o.Tail.N)
})

func graveEndeavorDrainLife(ctx *Context, x int) error {
	if x <= 0 {
		return nil
	}
	if err := (EachOpponentLosesLife{N: x}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Amount: x}.Apply(ctx)
}
