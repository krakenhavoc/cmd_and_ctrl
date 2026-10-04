package effects

import (
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
// itself" (Emergent Ultimatum) passes game.LapseStaysInExile: the window
// closes on the caster's next priority pass and an uncast card stays in
// exile. An ability that says "until end of turn, you may cast that
// card" (Nicol Bolas) passes "" and keeps the whole turn, so the card
// can still be cast in the second main phase.

// grantFreeCasts stamps a free-cast permission on each of `ids` that is
// still in exile. `timing` is TimingNormal for an ability that offers
// the cast "this turn" (the card keeps its own timing), and TimingFlash
// for a spell's resolution that casts "as part of" itself (CR 608.2g).
func grantFreeCasts(g *game.Game, controller, source uuid.UUID, sourceName string,
	timing game.GrantTiming, lapse game.PermissionLapse, ids []uuid.UUID) {
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
	}
}
