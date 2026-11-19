package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b08_helpers.go — the pieces of Death Cloud (S82 BNS batch 8).

// deathCloudDiscardThen is "sacrifices X creatures of their choice", run
// once the last discard has landed. A keyed discard run hands its
// continuation only a carry of IDs, so X rides on the carry's length: the
// caster first, then one uuid.Nil per point of X.
var deathCloudDiscardThen = RunThen("prompt-run/death-cloud-sacrifice-creatures", func(ctx *Context, r game.RunDone) error {
	x := len(r.Carry) - 1
	if x <= 0 {
		return nil
	}
	return ctx.PlayersSacrificeThen(deathCloudSeatsOwing(ctx.Game, x),
		sacrificeSpec("a creature", Creature()), "Death Cloud — sacrifice X creatures",
		deathCloudSacrificeLands, ctx.RunData(r.Carry[:1], x))
})

// deathCloudSacrificeLands is "then sacrifices X lands of their choice",
// run once the creatures have left. X is the run's first number.
var deathCloudSacrificeLands = RunThen("prompt-run/death-cloud-sacrifice-lands", func(ctx *Context, r game.RunDone) error {
	if len(r.Nums) == 0 || r.Nums[0] <= 0 {
		return nil
	}
	ctx.Game.PlayersSacrificeForEffect(ctx.Source(), deathCloudSeatsOwing(ctx.Game, r.Nums[0]),
		sacrificeSpec("a land", Land()), "Death Cloud — sacrifice X lands")
	return nil
})

// deathCloudSeatsOwing lists every living seat x times: a name listed N
// times owes N permanents (PlayersSacrificeThen).
func deathCloudSeatsOwing(g *game.Game, x int) []uuid.UUID {
	var out []uuid.UUID
	for _, seat := range apnapLivingSeats(g) {
		for i := 0; i < x; i++ {
			out = append(out, seat)
		}
	}
	return out
}

// deathCloudResolve is the printed body: life, then discards, then
// creatures, then lands, each finishing before the next (CR 608.2c).
func deathCloudResolve(_ *game.StackItem, ctx *Context) error {
	x := ctx.X()
	if x <= 0 {
		return nil
	}
	for _, seat := range apnapLivingSeats(ctx.Game) {
		if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), seat, -x); err != nil {
			return err
		}
	}
	carry := make([]uuid.UUID, 1, x+1)
	carry[0] = ctx.Controller()
	for i := 0; i < x; i++ {
		carry = append(carry, uuid.Nil)
	}
	return ctx.Game.EachPlayerDiscardsThenKeyedForEffect(uuid.Nil,
		game.DiscardPrompt{Source: ctx.Source(), N: x, Question: "Death Cloud — discard X cards"},
		deathCloudDiscardThen, carry)
}

// s82BNSB08MillTargetPlayer is "target player mills a card" (Duskmantle,
// House of Shadow).
func s82BNSB08MillTargetPlayer(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	victim, ok := firstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	return MillCards{Player: victim, N: 1}.Apply(ctx)
}
