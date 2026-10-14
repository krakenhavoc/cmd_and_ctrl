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
// and Kaervek, the Punisher with a price and a follow-up; Spellweaver
// Volute and Uldaros Theorix are the same call with a budget added when
// they ship. It is NOT
// CopySpellForEffect (a copy of a spell already on the stack, which is
// cast already) and NOT CreateTokenCopy (a permanent on the battlefield).
//
// The copies are offered to Player together, with no total cap, and Player
// casts them in the order they like — each is its own grant, so casting one
// never touches another. An instant or sorcery copy ceases to exist as it
// resolves; a permanent copy becomes a token (CR 707.10b). Every copy not
// cast when the window closes, on Player's next priority pass, ceases to
// exist too (CR 704.5e).

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
	// at 0, CR 107.3b). False is "you may cast the copy": its printed
	// mana cost is paid.
	Free bool

	// FollowUp is a registered body (game.RegisterCastFollowUp) run when a
	// copy is cast because of this offer — "if you do, you lose 2 life".
	FollowUp game.CastFollowUpKey
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
	}
	if src, ok := ctx.Game.LookupCardForEffect(ctx.Source()); ok {
		offer.SourceName = src.Name
	}
	if c.Free {
		offer.Cost = "{0}"
	}
	ctx.Game.CopyCardsAndOfferCastsForEffect(offer, c.Cards)
	return nil
}
