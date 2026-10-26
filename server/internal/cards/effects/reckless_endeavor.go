package effects

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reckless Endeavor — "Roll two d12 and choose one result. Reckless
// Endeavor deals damage equal to that result to each creature. Then create
// a number of Treasure tokens equal to the other result."
func init() {
	Register(Spec{
		OracleID: "928bf0d9-89be-462d-aeb7-74a65e95535c", Name: "Reckless Endeavor", Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return RollDice{Sides: 12, N: 2, Then: recklessEndeavorRolled}.Apply(ctx)
		},
	})
}

// recklessEndeavorRolled is "choose one result", after the two d12 (#4307:
// a replaced roll waits for its chooser first).
var recklessEndeavorRolled = DiceTail("dice/reckless-endeavor-rolled", func(ctx *Context, rolls []int) error {
	if len(rolls) < 2 {
		return nil
	}
	item := ctx.Item
	controller, source := item.Controller, item.SourceCardID
	if rolls[0] == rolls[1] {
		return recklessEndeavorApply(ctx.Game, source, controller, rolls[0], rolls[1])
	}
	// Keyed (#4298): the two rolls ride the prompt as its word.
	return MayChoice{
		Question: "Reckless Endeavor — choose the damage result",
		YesLabel: fmt.Sprintf("%d damage, %d Treasures", rolls[0], rolls[1]),
		NoLabel:  fmt.Sprintf("%d damage, %d Treasures", rolls[1], rolls[0]),
		Then:     recklessEndeavorChoice,
		Word:     strconv.Itoa(rolls[0]) + " " + strconv.Itoa(rolls[1]),
	}.Apply(ctx)
})

// recklessEndeavorChoice is "choose one result": a yes is the first roll as
// the damage and the second as the Treasures, a no the other way round.
var recklessEndeavorChoice = MayAnswer("confirm/reckless-endeavor-result", func(ctx *Context, r game.ConfirmAnswer) error {
	first, second, ok := strings.Cut(r.Word, " ")
	a, errA := strconv.Atoi(first)
	b, errB := strconv.Atoi(second)
	if !ok || errA != nil || errB != nil {
		return nil
	}
	if !r.Accepted {
		a, b = b, a
	}
	return recklessEndeavorApply(ctx.Game, r.Source, ctx.Controller(), a, b)
})

// recklessEndeavorApply is the rest of the card: damage equal to the chosen
// result to each creature, then Treasures equal to the other.
func recklessEndeavorApply(g *game.Game, source, controller uuid.UUID, damage, treasures int) error {
	var targets []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() {
			targets = append(targets, c.InstanceID)
		}
	}
	return g.DealDamageEachThenForEffect(source, targets, damage, func(g *game.Game, _ int) error { return treasureTokens(g, controller, treasures) })
}
