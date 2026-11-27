package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b706_helpers.go — the Urza's Destiny "reveal any number of <colour>
// cards in your hand" family (S82 #5621): the Scents and the Seers.

// OldB706RevealAnyNumber is the shared opening of those cards: "Reveal any number
// of <colour> cards in your hand." and then `then` with the number revealed.
// `color` is the colour letter, `pickKey` the card's own "choose-cards/<card>" key
// (an on-disk identity, never renamed or reused).
//
// The reveal is made as the spell or ability resolves (CR 608.2d): the
// controller picks any number of the colour-matching cards in their hand
// (colours as each card has them in hand, CR 105.2), they are revealed to the
// table (CR 701.20a) and `then` runs with how many. `then` runs against the asking
// item, so it reads its targets. With no card of the colour in hand nothing is
// asked, nothing is revealed and `then` is not run: the number is 0, which each
// card's consequence treats as nothing. `then` must capture nothing.
func OldB706RevealAnyNumber(name, color, pickKey string, then func(ctx *Context, n int) error) func(item *game.StackItem, ctx *Context) error {
	colorName := game.ColorName(color)
	pick := AskedPick(pickKey, func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) == 0 {
			return nil
		}
		ctx.Game.RevealForEffect(game.RevealSpec{
			Player: ctx.Controller(),
			Source: ctx.Source(),
			Reason: name + " — revealed " + colorName + " cards",
			Cards:  r.Picked,
		})
		return then(ctx, len(r.Picked))
	})
	return func(item *game.StackItem, ctx *Context) error {
		p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
		if p == nil || p.Hand == nil {
			return nil
		}
		var cards []uuid.UUID
		for _, c := range p.Hand.Cards {
			if c.HasColorWord(color) {
				cards = append(cards, c.InstanceID)
			}
		}
		if len(cards) == 0 {
			return nil
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: name + " — reveal any number of " + colorName + " cards in your hand",
			Cards:    cards,
			Min:      0,
			Max:      len(cards),
			LiveZone: game.ZoneHand,
			ThenKey:  pick,
			Asked:    game.NewAskedItem(item),
		})
		return nil
	}
}

// OldB706AsEffect adapts a spell-shaped body to an activated ability's Effect.
func OldB706AsEffect(f func(item *game.StackItem, ctx *Context) error) Effect {
	return func(g *game.Game, item *game.StackItem) error { return f(item, NewContext(g, item)) }
}

// OldB706PumpByRevealed is "target creature gets +X/+X or -X/-X until end of
// turn, where X is the number of cards revealed this way", `sign` being +1 or -1.
func OldB706PumpByRevealed(name string, sign int) func(ctx *Context, n int) error {
	return func(ctx *Context, n int) error {
		t, ok := ctx.ClauseTarget(0)
		if !ok {
			return nil
		}
		return BoostUntilEOT{Target: t.ID, Power: sign * n, Toughness: sign * n, Label: name}.Apply(ctx)
	}
}

// OldB706GainTwoPerRevealed is "You gain 2 life for each card revealed this way."
func OldB706GainTwoPerRevealed(ctx *Context, n int) error {
	return GainLife{Amount: 2 * n}.Apply(ctx)
}

// OldB706CounterUnlessPaysPerRevealed is "Counter target spell unless its
// controller pays {1} for each card revealed this way."
func OldB706CounterUnlessPaysPerRevealed(name string) func(ctx *Context, n int) error {
	return func(ctx *Context, n int) error {
		t, ok := ctx.ClauseTarget(0)
		if !ok {
			return nil
		}
		return CounterUnlessPaid{
			StackID:  t.ID,
			Cost:     fmt.Sprintf("{%d}", n),
			Question: fmt.Sprintf("%s — pay {%d} or your spell is countered", name, n),
		}.Apply(ctx)
	}
}
