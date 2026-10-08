package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deeproot Historian — Creature — Merfolk Druid {3}{G}, 3/3:
//
//	"Merfolk and Druid cards in your graveyard have retrace. (You may
//	 cast cards with retrace from your graveyard by discarding a land
//	 card in addition to paying their other costs.)"
//
// A STANDING cast permission (ADR 0066, 2026-10-08 amendment, #2550),
// declared on Spec.CastPermissions and derived from the battlefield on
// every query, so it lasts exactly as long as the Historian does and
// two of them compose. It is Six's grant (six.go) with a different
// filter and no timing clause: nothing here says "during your turn",
// so the card's OWN timing is the only one in force (a Merfolk creature
// card is a main-phase cast, a flash one is castable at instant speed).
//
// "Merfolk AND Druid cards" is a union — a card with either creature
// type qualifies, a Merfolk Wizard included — which is
// PermissionFilter.CreatureTypesAny. The Historian counts: it is a
// Merfolk Druid card and a copy of it in the graveyard retraces. A
// retraced creature is a permanent spell, so it resolves onto the
// battlefield and, dying again, can be retraced again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b8d82fe-571d-4fd6-8d0c-77b7e5ef3e23",
		Name:         "Deeproot Historian",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:            game.ZoneGraveyard,
			Filter:          game.PermissionFilter{CreatureTypesAny: [2]string{"Merfolk", "Druid"}},
			AltCostKey:      game.AltCostKeyRetrace,
			DiscardLandCard: true,
			Label:           "Retrace — discard a land card (Deeproot Historian)",
		}},
	})
}
