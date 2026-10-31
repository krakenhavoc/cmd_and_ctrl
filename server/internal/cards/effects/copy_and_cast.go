package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// copy_and_cast.go — CopyAndCast: "copy each of those cards. You may cast
// any number of the copies" (CR 707.12; ADR 0066's 2026-10-13 card-copy
// amendment, #3258).
//
// It is the card-facing door onto game.CopyCardsAndOfferCastsForEffect,
// the one place the engine makes a copy of a CARD and offers a cast of it.
// Mizzix's Mastery and Arcane Bombardment use it for a set of any size,
// Kaervek, the Punisher with a price and a follow-up, and Uldaros Theorix
// with a total mana value budget (#2017), and Spellweaver Volute with a
// follow-up about the card it copied (#3157). It is NOT CopySpellForEffect (a copy of a spell
// already on the stack, which is cast already) and NOT CreateTokenCopy (a
// permanent on the battlefield).
//
// The copies are offered to Player together, and Player casts them in the
// order they like, DURING the resolution (CR 707.12, CR 608.2g; ADR 0066's
// 2026-10-22 amendment, #2947): a resolution_cast prompt holds the
// resolution open until Player has cast the ones they want, each becoming
// the topmost object on the stack with nobody receiving priority in
// between. With no total each is its own grant, so casting one never
// touches another; with TotalManaValueAtMost they are ONE grant that every
// cast spends from. An instant or sorcery copy ceases to exist as it
// resolves; a permanent copy becomes a token (CR 707.10b). Every copy not
// cast ceases to exist too (CR 704.5e).

// CopyAndCast makes a copy of every card in Cards and offers Player a cast
// of each.
type CopyAndCast struct {
	// Cards are the originals, in exile or a graveyard. The order is the
	// order the copies are listed in; the cast order is Player's.
	Cards []uuid.UUID

	// Player creates, owns and casts the copies; zero is the resolving
	// item's controller.
	Player uuid.UUID

	// Free is "without paying its mana cost" ({0}, which also locks an X
	// at 0, CR 107.3b). False with an empty Cost is "you may cast the
	// copy": its printed mana cost is paid.
	Free bool

	// Cost is "you may cast the copy by paying {3} rather than paying its
	// mana cost" (CR 118.9, Blue Mage's Cane, #3783): the price paid
	// INSTEAD of the printed mana cost, written as a mana cost. Free wins
	// when both are set, since Free is Cost "{0}".
	Cost string

	// FollowUp is a registered body (game.RegisterCastFollowUp) run when a
	// copy is cast because of this offer — "if you do, you lose 2 life".
	FollowUp game.CastFollowUpKey

	// FollowUpObject is handed to FollowUp as CastFollowUp.Object, for an
	// "if you do" about an object the copy spell does not name
	// (Spellweaver Volute's "exile the enchanted card", #3157).
	FollowUpObject game.ObjectRef

	// LastKnown copies a card the effect saw in a graveyard even after it
	// has left (CR 608.2h; game.CardCopyOffer.LastKnown).
	LastKnown bool

	// TotalManaValueAtMost is "any number of spells with total mana value
	// N or less from among the copies" (Uldaros Theorix, #2017): the casts
	// share one running total, each spending its mana value (X at 0, as a
	// free cast locks it) and refused once it would pass N. Nil is no
	// total, the ordinary "any number of the copies".
	TotalManaValueAtMost *int

	// AtMostCasts is "you may cast up to N of the copies" (Baron Helmut
	// Zemo's boast, N = 3): the copies share one permission that allows N
	// casts (game.CastPermission.CastsLeft); the copies not cast cease to
	// exist. Zero is no limit. The copies of one call sit in one zone.
	AtMostCasts int
}

// Apply makes the copies and the offers.
func (c CopyAndCast) Apply(ctx *Context) error {
	player := c.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	offer := game.CardCopyOffer{
		Player:   player,
		Source:   ctx.Source(),
		Timing:   game.TimingFlash,
		Lapse:    game.LapseStaysInExile,
		FollowUp: c.FollowUp,

		DuringResolution: true,

		FollowUpObject: c.FollowUpObject,
		LastKnown:      c.LastKnown,
		MaxCasts:       c.AtMostCasts,
	}
	if c.TotalManaValueAtMost != nil {
		n := *c.TotalManaValueAtMost
		offer.ManaValueBudget = &n
	}
	if src, ok := ctx.Game.LookupCardForEffect(ctx.Source()); ok {
		offer.SourceName = src.Name
	}
	switch {
	case c.Free:
		offer.Cost = "{0}"
	case c.Cost != "":
		offer.Cost = c.Cost
	}
	ctx.Game.CopyCardsAndOfferCastsForEffect(offer, c.Cards)
	return nil
}

// exileCopyCastThenExileSelf is the body Mizzix's Mastery and Psionic
// Ritual share: exile `ids` through the CR 614 window, copy only the
// cards that really reached exile with a free-cast offer each (CR
// 707.12), then exile the spell itself ("Exile <this>").
func exileCopyCastThenExileSelf(item *game.StackItem, ctx *Context, ids []uuid.UUID) error {
	return ctx.Game.ExileCardsKeyedForEffect(ids, NewContext(ctx.Game, item).ItemTail(exileCopyCastThenExileSelfTail))
}

// exileCopyCastThenExileSelfTail casts copies of what reached exile, then exiles the source (#4253 slice 9).
var exileCopyCastThenExileSelfTail = CardTail("copy-and-cast/cast-copies-then-exile-self", func(ctx *Context, o game.TailOutcome) error {
	g, item, exiled := ctx.Game, ctx.Item, o.IDs

	ctx = NewContext(g, item)
	if err := (CopyAndCast{Cards: exiled, Free: true}).Apply(ctx); err != nil {
		return err
	}
	return ExileTarget{Target: item.SourceCardID}.Apply(ctx)
})
