package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b23_helpers.go — the shared pieces of Standard-set batch 23
// (Duskmourn, #4466): the Rakdos "sacrifice another creature or
// enchantment" cost and the effects its cards share.

// b23SacrificeAnotherCreatureOrEnchantment is the cost clause "Sacrifice
// another creature or enchantment" (Popular Egotist, Diversion
// Specialist, Sawblade Skinripper): one permanent the activator controls
// that is a creature or an enchantment, never the source itself (CR
// 109.1, "another"). It is a cost (CR 602.2), paid as the ability is
// activated, and it counts as a sacrifice for "whenever you sacrifice".
func b23SacrificeAnotherCreatureOrEnchantment() game.AbilityCost {
	return SacrificeAnotherN(1, "another creature or enchantment", Or(Creature(), Enchantment()))
}

// b23IndestructibleAndTapThis is "This creature gains indestructible until
// end of turn. Tap it." as an activated ability's effect, the source being
// the creature (Popular Egotist). Both halves are the resolving ability's
// own, so it does both even if the creature was already tapped.
func b23IndestructibleAndTapThis(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if err := (GrantKeywordUntilEOT{
			Target:   item.SourceCardID,
			Keywords: []string{"indestructible"},
			Label:    label,
		}).Apply(ctx); err != nil {
			return err
		}
		return TapTarget{Target: item.SourceCardID}.Apply(ctx)
	}
}

// b23YouSacrificedAPermanent is "whenever you sacrifice a permanent": one
// EventSacrifice per permanent, actor the source's controller. It fires
// while the permanent is still on the battlefield, so sacrificing the
// source itself triggers its own ability too (the Popular Egotist ruling).
func b23YouSacrificedAPermanent(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Actor == source.Controller && ev.CardID != uuid.Nil
}

// b23PermanentsYouSacrificedThisTurn is "the number of permanents you
// sacrificed this turn" for the resolving item's controller: the turn
// tally counts every one, tokens included (CR 704.5d), and an opponent's
// sacrifices are not in it (Sawblade Skinripper).
func b23PermanentsYouSacrificedThisTurn(ctx *Context) int {
	return ctx.Game.TurnTallyFor(ctx.Controller()).PermanentsSacrificed
}

// b23HasID reports whether ids holds id.
func b23HasID(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}
