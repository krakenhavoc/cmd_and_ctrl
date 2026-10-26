package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// free_cast_grants.go — "you may cast <those exiled cards> without
// paying their mana costs", for the resolutions that print the clause
// over a set they just exiled (Emergent Ultimatum, Nicol Bolas,
// God-Pharaoh).
//
// The free cast is a GRANT, not an inline cast: ADR 0066's posture on
// every "you may cast it" a resolution offers (cascade, Malcolm). The
// price is spelled "{0}" because an empty Cost means "pay the printed
// cost", and it locks X at 0 (CR 107.3b). CastOnly, because the text
// says cast. Duration is until end of turn.
//
// `lapse` is the window's other end. A resolution that casts "as part of
// itself" (Emergent Ultimatum) passes game.LapseStaysInExile: the cast is
// made during the resolution (CR 608.2g, #4250), through the
// resolution_cast prompt over the cards, each its own answer until the
// player declines or none is left, and an uncast card stays in exile.
// An ability that says "until end of turn, you may cast that card" (Nicol
// Bolas) passes "" and keeps the whole turn, so the card can still be
// cast in the second main phase; it asks no prompt.

// grantFreeCasts stamps a free-cast permission on each of `ids` that is
// still in exile. `timing` is TimingNormal for an ability that offers
// the cast "this turn" (the card keeps its own timing), and TimingFlash
// for a spell's resolution that casts "as part of" itself (CR 608.2g).
func grantFreeCasts(g *game.Game, controller, source uuid.UUID, sourceName string,
	timing game.GrantTiming, lapse game.PermissionLapse, ids []uuid.UUID) {
	var granted []uuid.UUID
	defer func() {
		if lapse != "" {
			offerResolutionCastOver(g, controller, source, sourceName, game.MayCastKeywordFreeCast, declineLeaveInExile, granted...)
		}
	}()
	for _, id := range ids {
		z := g.FindCardZoneForEffect(id)
		c, ok := g.LookupCardForEffect(id)
		if !ok || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:      controller,
			Zone:        game.ZoneExile,
			Duration:    g.UntilEndOfTurnDuration(),
			Cost:        "{0}",
			Timing:      timing,
			CastOnly:    true,
			LapseOnPass: lapse,
			Source:      source,
			SourceName:  sourceName,
			Label:       sourceName + " — cast it without paying its mana cost",
		}, []game.Card{c})
		granted = append(granted, id)
	}
}

// The decline labels the one-off offers use (#4250): where a card that is
// not cast stays.
const (
	declineLeaveInExile     = "Leave it in exile"
	declineLeaveInGraveyard = "Leave it in the graveyard"
)

// offerResolutionCastOver asks `controller` to cast `ids` during the
// resolution offering them (CR 608.2g, #4250): the resolution_cast prompt
// over cards a permission the caller has just stamped names. Each card is
// its own answer while its permission lasts; a permission over several
// cards that is spent by one cast ("cast a spell from among them") closes
// the offer with that cast. The decline lapses the permissions, leaving
// the cards where they are. Outside an open resolution nothing is asked
// and the permissions alone are the window, closing on the holder's next
// priority pass.
func offerResolutionCastOver(g *game.Game, controller, source uuid.UUID, name, keyword, decline string, ids ...uuid.UUID) {
	accept := "Cast it free"
	if keyword == game.MayCastKeywordPaidCast {
		accept = "Cast it"
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       name,
		Cards:        ids,
		Keyword:      keyword,
		AcceptLabel:  accept,
		DeclineLabel: decline,
	})
}

// exileWithPermissionThenOffer is game.ExileCardWithPermissionForEffect
// followed, once the card has landed in exile, by the resolution_cast
// offer over it (#4250). A card that does not reach exile (its owner sent
// a commander to the command zone) is offered nothing. An empty name heads
// the offer with the exiled card's own.
func exileWithPermissionThenOffer(g *game.Game, cardID, source uuid.UUID, perm game.CastPermission, name, keyword string) error {
	if z := g.FindCardZoneForEffect(cardID); z != nil && z.Kind == game.ZoneExile {
		exiledWithPermissionOffer(g, cardID, source, perm, name, keyword)
		return nil
	}
	// Keyed (#4253 slice 10a): the permission, the offer's heading and its
	// keyword ride the exile's tail.
	return g.ExileCardKeyedForEffect(cardID, game.KeyedTail{Ref: exiledThenOffer, Source: source, Permission: &perm,
		Word: keyword + freeCastWordSep + name})
}

// freeCastWordSep separates exiledThenOffer's keyword from its heading in
// the tail's Word.
const freeCastWordSep = "\x1f"

// exiledThenOffer is exileWithPermissionThenOffer's continuation, told the
// card when it reached exile.
var exiledThenOffer = game.RegisterTail("exile/grant-permission-then-offer", func(g *game.Game, o game.TailOutcome) error {
	if len(o.IDs) != 1 || o.Tail.Permission == nil {
		return nil
	}
	keyword, name, _ := strings.Cut(o.Tail.Word, freeCastWordSep)
	exiledWithPermissionOffer(g, o.IDs[0], o.Tail.Source, *o.Tail.Permission, name, keyword)
	return nil
})

// exiledWithPermissionOffer grants perm over the exiled card and offers the
// cast over it.
func exiledWithPermissionOffer(g *game.Game, cardID, source uuid.UUID, perm game.CastPermission, name, keyword string) {
	perm.Zone = game.ZoneExile
	if !g.GrantCastPermissionOverCardForEffect(cardID, perm) {
		return
	}
	reason := name
	if c, found := g.LookupCardForEffect(cardID); found && reason == "" {
		reason = c.Name
	}
	offerResolutionCastOver(g, perm.Player, source, reason, keyword, declineLeaveInExile, cardID)
}
