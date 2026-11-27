package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b705_helpers.go — the Urza's Destiny "Scent of" instants (S82 #5621):
// "Reveal any number of <colour> cards in your hand. Target creature gets
// +X/+X or -X/-X until end of turn, where X is the number of cards revealed
// this way."

// OldB705ScentOfX builds one Scent's OnResolve. `color` is the colour letter of
// the cards that may be revealed, `sign` is +1 or -1, and `pickKey` is the card's
// own "choose-cards/<card>" key (an on-disk identity, never renamed or reused).
//
// The reveal is made as the spell resolves (CR 608.2d): the caster picks any
// number of the colour-matching cards in their hand (colours as each card has
// them in hand, CR 105.2), they are revealed to the table (CR 701.20a), and the
// target gets the pump or shrink for that many. A target that has left or become
// illegal is judged before this runs (CR 608.2b), so a fizzled Scent reveals
// nothing. With no card of the colour in hand nothing is asked and nothing is
// revealed (X is 0).
func OldB705ScentOfX(name, color string, sign int, pickKey string) func(item *game.StackItem, ctx *Context) error {
	colorName := game.ColorName(color)
	apply := func(ctx *Context, n int, target uuid.UUID) error {
		if n == 0 || target == uuid.Nil {
			return nil
		}
		return BoostUntilEOT{
			Target:    target,
			Power:     sign * n,
			Toughness: sign * n,
			Label:     name,
		}.Apply(ctx)
	}
	pick := AskedPick(pickKey, func(ctx *Context, r game.CardsChosen) error {
		t, ok := ctx.ClauseTarget(0)
		if !ok || len(r.Picked) == 0 {
			return nil
		}
		ctx.Game.RevealForEffect(game.RevealSpec{
			Player: ctx.Controller(),
			Source: ctx.Source(),
			Reason: name + " — revealed " + colorName + " cards",
			Cards:  r.Picked,
		})
		return apply(ctx, len(r.Picked), t.ID)
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
