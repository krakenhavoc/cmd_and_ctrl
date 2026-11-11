package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s49_helpers.go — shared bodies for the batch 49 cards (#5368):
// the tempting offer that asks each opponent in turn.

// cmdr2S49Tempting is a tempting offer whose opponent half is a plain
// effect on a player ("draw a card and create a 1/1 Rabbit"): the caster
// already did it, each opponent in turn order is asked whether to do it,
// and for each who says yes the caster does it again. The question is
// asked of one opponent at a time; the opponents still to ask ride the
// prompt as its carry.
type cmdr2S49Tempting struct {
	question string
	do       func(ctx *Context, player uuid.UUID) error
	answer   game.ConfirmThen
}

// newCmdr2S49Tempting registers the keyed answer. Call it from init.
func newCmdr2S49Tempting(key, question string, do func(ctx *Context, player uuid.UUID) error) *cmdr2S49Tempting {
	t := &cmdr2S49Tempting{question: question, do: do}
	t.answer = MayAnswer(key, func(ctx *Context, r game.ConfirmAnswer) error {
		if r.Accepted {
			if err := t.do(ctx, r.Chooser); err != nil {
				return err
			}
			if err := t.do(ctx, ctx.Controller()); err != nil {
				return err
			}
		}
		return t.offer(ctx, r.Carry)
	})
	return t
}

// resolve runs the caster's own half, then offers it to each opponent.
func (t *cmdr2S49Tempting) resolve(ctx *Context) error {
	if err := t.do(ctx, ctx.Controller()); err != nil {
		return err
	}
	return t.offer(ctx, ctx.Opponents())
}

func (t *cmdr2S49Tempting) offer(ctx *Context, opponents []uuid.UUID) error {
	if len(opponents) == 0 {
		return nil
	}
	return MayChoice{
		Player:   opponents[0],
		Question: t.question,
		Then:     t.answer,
		Carry:    opponents[1:],
	}.Apply(ctx)
}

// cmdr2S49Parley is the parley body shared by Ambassador K'Ehleyr, Selvala,
// Storm Fleet Negotiator and Innocuous Researcher: "each player reveals
// the top card of their library" (APNAP order, every player still in the
// game), the caller turns the number of nonland cards revealed into its
// reward, "then each player draws a card". A player with an empty library
// reveals nothing, and the draw from it is the CR 704.5b loss as usual.
// `reward` runs between the reveals and the draws, so the printed word
// order ("for each nonland card revealed this way ... Then each player
// draws") holds; K'Ehleyr's reward comes after the draws, as printed there.
func cmdr2S49Parley(g *game.Game, source uuid.UUID, label string, reward func(nonland int) error) error {
	players := apnapPlayers(g)
	nonland := 0
	for _, p := range players {
		for _, id := range g.RevealTopOfLibraryForEffect(p, source, 1, label) {
			if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
				nonland++
			}
		}
	}
	if reward != nil {
		if err := reward(nonland); err != nil {
			return err
		}
	}
	for _, p := range players {
		if err := g.DrawNForEffect(p, 1); err != nil {
			return err
		}
	}
	return nil
}
