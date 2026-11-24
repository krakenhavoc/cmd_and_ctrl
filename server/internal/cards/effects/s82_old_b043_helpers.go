package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b043_helpers.go — shared pieces for the Strixhaven: School of Mages
// batch 043 (S82, #5502).

// OldB43ExtractAndVictimDraws is the extraction family's "search <victim>'s
// graveyard, hand, and library for any number of cards with the same name as
// <the spell or permanent> and exile them. That player shuffles, then draws a
// card for each card exiled from their hand this way" (The End, Test of
// Talents). The victim's hand is recorded as the search begins, and the draw
// counts the found cards that were in it, one card each.
func OldB43ExtractAndVictimDraws(ctx *Context, card string, victim uuid.UUID, names []string) error {
	if victim == uuid.Nil {
		return nil
	}
	var hand []uuid.UUID
	if p := ctx.Game.PlayerByIDForEffect(victim); p != nil && p.Hand != nil {
		for _, hc := range p.Hand.Cards {
			hand = append(hand, hc.InstanceID)
		}
	}
	return SearchLibrary{
		Player:               ctx.Controller(),
		LibraryOwner:         victim,
		Predicate:            SharesANameWith(names),
		Dest:                 game.ZoneExile,
		Unbounded:            true,
		AlsoGraveyardAndHand: true,
		Shuffle:              true,
		Reason:               card + " — search their graveyard, hand and library for cards with the same name",
		ThenTail:             WithCarry(ctx.ItemTail(oldB43VictimDraws), append([]uuid.UUID{victim}, hand...)...),
	}.Apply(ctx)
}

// oldB43VictimDraws is the last sentence: the victim draws a card for each
// found card that was in their hand when the search began. An on-disk key.
var oldB43VictimDraws = CardTail("old-b043/extraction-victim-draws", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	victim, inHand := o.Tail.Carry[0], map[uuid.UUID]bool{}
	for _, id := range o.Tail.Carry[1:] {
		inHand[id] = true
	}
	n := 0
	for _, id := range o.IDs {
		if inHand[id] {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return DrawCards{Player: victim, N: n}.Apply(ctx)
})

// OldB43RevealUntilAndOfferFreeCast is "reveal cards from the top of your
// library until you reveal a card that <hit> admits. You may cast that card
// without paying its mana cost. Put all revealed cards not cast this way on
// the bottom of your library in a random order" (Plargg, Dean of Chaos). The
// walk only reveals: the hit stays in the library while the cast is offered
// over it (Aetherworks Marvel's resolution_cast prompt, one free spell), and
// afterwards every revealed card still in a library, the declined hit and a
// hit that cannot be cast included, goes to the bottom in a random order
// (LookedAtRestToBottom). A library with no hit is revealed whole and bottomed.
func OldB43RevealUntilAndOfferFreeCast(ctx *Context, name string, hit func(game.Card) bool) error {
	g, you, source := ctx.Game, ctx.Controller(), ctx.Source()
	run, found := RevealUntil(ctx, you, hit, name+" — reveal until a card you may cast")
	if len(run) == 0 {
		return nil
	}
	if found != uuid.Nil {
		perm := FreeCastSpell{}.On(g.FreeCastFromLibraryPermission(you, source, name))
		if eligible := FreeCastCandidates(g, &perm, []uuid.UUID{found}); len(eligible) > 0 {
			perm.LapseOnPass = game.LapseStaysInHand
			if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
				Chooser:      you,
				Source:       source,
				Reason:       name,
				Cards:        eligible,
				Permission:   &perm,
				Keyword:      game.MayCastKeywordFreeCast,
				AcceptLabel:  "Cast it free",
				DeclineLabel: "Put them on the bottom",
				Then:         LookedAtRestToBottom,
				Carry:        run,
			}) {
				return nil
			}
		}
	}
	return g.PutOnBottomInRandomOrderForEffect(you, game.ZoneLibrary, CardsStillInALibrary(g, run))
}
