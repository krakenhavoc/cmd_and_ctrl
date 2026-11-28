package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b852_helpers.go — shared pieces of S82 OLD batch 852 (Fallen Empires,
// #5637).

// S82OldB852ShroudFreezeAndTap is the effect of the Homarids' "{U}: This
// creature gains shroud until end of turn and doesn't untap during your next
// untap step. Tap it." (Homarid Warrior, Deep Spawn), in the printed order: the
// shroud, the freeze on the object it is now (a creature that left and came back
// is a new object, CR 400.7), then the tap. `name` labels the effects.
func S82OldB852ShroudFreezeAndTap(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if err := ThisGainsKeywordUntilEndOfTurn("shroud", name+" — shroud until end of turn")(g, item); err != nil {
			return err
		}
		if err := (DoesntUntapNextUntapStep{
			Targets: []uuid.UUID{item.SourceCardID},
			Player:  item.Controller,
			Label:   name + " — doesn't untap during your next untap step",
		}).Apply(ctx); err != nil {
			return err
		}
		return TapTarget{Target: item.SourceCardID}.Apply(ctx)
	}
}

// S82OldB852SacrificeUnlessYouMill is the upkeep tax of "At the beginning of your
// upkeep, sacrifice this creature unless you mill two cards" (Deep Spawn): a
// two-way choice when the library holds `n` cards to mill, and an unconditional
// sacrifice when it does not (a cost can't be paid in part, CR 118.3). Milling is
// not one of the pay-unless payments, so it is a PickOption between the two, as
// the sacrifice taxes are, answered by a keyed continuation.
func S82OldB852SacrificeUnlessYouMill(name string, n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		p := g.PlayerByIDForEffect(item.Controller)
		if p == nil || p.Library == nil || len(p.Library.Cards) < n {
			return SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
		}
		return PickOption{
			Question: name + " — mill two cards, or sacrifice " + name,
			Options: []game.ChoiceOption{
				{Label: "Mill two cards"},
				{Label: "Sacrifice " + name},
			},
			ThenKey: s82OldB852MillOrSacrifice,
		}.Apply(ctx)
	}
}

// s82OldB852MillOrSacrifice is the answer to the choice above: index 0 mills two
// cards, anything else sacrifices the permanent. An on-disk key: never renamed,
// never reused.
var s82OldB852MillOrSacrifice = OptionPickThen("option-pick/s82-old-b852-sacrifice-unless-you-mill-two", func(ctx *Context, r game.OptionPicked) error {
	if r.Index == 0 {
		return MillCards{N: 2}.Apply(ctx)
	}
	return SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
})
