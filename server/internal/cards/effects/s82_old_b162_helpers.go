package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b162_helpers.go — shared pieces of Modern Horizons batch 162
// (#5519, S82).

// S82OldB162BottomExiled (shared with Ryan Sinclair) puts the cards of `ids` that are still in exile on
// the bottom of the player's library in a random order; one that was cast has
// left exile and stays where it went.
func S82OldB162BottomExiled(g *game.Game, player uuid.UUID, ids []uuid.UUID) error {
	var rest []uuid.UUID
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			rest = append(rest, id)
		}
	}
	if len(rest) == 0 {
		return nil
	}
	return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneExile, rest)
}

// s82OldB162CastRest is the casts answered: whatever was cast, the exiled
// cards that were not go to the bottom (Carry is every card exiled). An
// on-disk key.
var s82OldB162CastRest = game.RegisterChooseCardsThen("resolution-cast/s82-old-b162-exiled-rest-to-bottom",
	func(g *game.Game, r game.CardsChosen) error {
		return S82OldB162BottomExiled(g, r.Chooser, r.Carry)
	})

// S82OldB162CastUpToTwoSorceriesThenBottom is Collected Conjuring's whole
// clause over the cards just exiled: the caster may cast up to two sorcery
// spells with mana value 3 or less from among them without paying their mana
// costs (one permission over them with two casts to spend, offered during the
// resolution, CR 608.2g), and the exiled cards not cast go to the bottom of the
// library in a random order, whatever was answered.
func S82OldB162CastUpToTwoSorceriesThenBottom(g *game.Game, controller, source uuid.UUID, name string, exiled []uuid.UUID) error {
	var offered []game.Card
	var ids []uuid.UUID
	for _, id := range exiled {
		z := g.FindCardZoneForEffect(id)
		c, ok := g.LookupCardForEffect(id)
		if !ok || z == nil || z.Kind != game.ZoneExile || !c.IsSorcery() || c.ManaValue() > 3 {
			continue
		}
		offered = append(offered, c)
		ids = append(ids, id)
	}
	if len(offered) == 0 {
		return S82OldB162BottomExiled(g, controller, exiled)
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      controller,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		CastsLeft:   2,
		AnyNumber:   true,
		LapseOnPass: game.LapseStaysInExile,
		Source:      source,
		SourceName:  name,
		Label:       name + " — cast up to two sorcery spells with mana value 3 or less from among the exiled cards without paying their mana costs",
	}, offered)
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       name,
		Cards:        ids,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put the rest on the bottom",
		Then:         s82OldB162CastRest,
		Carry:        exiled,
	}) {
		return nil
	}
	return S82OldB162BottomExiled(g, controller, exiled)
}
