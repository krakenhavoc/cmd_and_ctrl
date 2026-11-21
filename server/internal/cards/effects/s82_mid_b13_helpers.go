package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mid_b13_helpers.go — shared bodies for Midnight Hunt Commander batch 13
// (#5497): the "choose creatures with different powers" picks of Sigarda's
// Vanguard, Sigardian Zealot and Celestial Judgment.

// midB13DifferentPowers is "creatures with different powers": no two picked
// creatures share a power, read through the layers and unclamped (a power of
// -1 differs from 0). On-disk behaviour of the picks below; never captures
// game state.
func MidB13DifferentPowers(picked []game.Card) bool {
	seen := make(map[int]bool, len(picked))
	for _, c := range picked {
		p := c.PowerForComparison()
		if seen[p] {
			return false
		}
		seen[p] = true
	}
	return true
}

// midB13CreaturesOnBattlefield lists every creature on the battlefield, any
// controller, in battlefield order, through the layers.
func MidB13CreaturesOnBattlefield(g *game.Game) []game.Card {
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() {
			out = append(out, c)
		}
	}
	return out
}

// midB13DistinctPowers is the number of different powers among cs.
func MidB13DistinctPowers(cs []game.Card) int {
	seen := make(map[int]bool, len(cs))
	for _, c := range cs {
		seen[c.PowerForComparison()] = true
	}
	return len(seen)
}

// midB13IDs is the instance IDs of cs, in order.
func midB13IDs(cs []game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(cs))
	for _, c := range cs {
		ids = append(ids, c.InstanceID)
	}
	return ids
}

// midB13ChooseDifferentPowers asks the controller to choose between min and
// max of the creatures on the battlefield, with different powers (the set
// rule refuses two of one power, so a table waiting on the answer is a
// restore point and a bot is never offered a set the rule would refuse).
// nums are handed to the continuation as CardsChosen.Nums and carry what was
// read before the choice.
func MidB13ChooseDifferentPowers(ctx *Context, creatures []game.Card, min, max int, question string, then game.ChooseCardsThen, nums ...int) {
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     ctx.Source(),
		Question:   question,
		Cards:      midB13IDs(creatures),
		Min:        min,
		Max:        max,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    then,
		CarryNums:  nums,
		Carry:      midB13IDs(creatures),
		Asked:      game.NewAskedItem(ctx.Item),
	})
}
