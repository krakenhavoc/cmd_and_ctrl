package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dream Halls — Enchantment {3}{U}{U} (Stronghold, S82 OLD b750, #5627):
//
//	"Rather than pay the mana cost for a spell, its controller may discard a
//	 card that shares a color with that spell."
//
// The granted alternative cost of ADR 0118 (CR 118.9) on every player's spells,
// not just its controller's (#6214): the discard is one card from hand that
// shares a colour with the spell being cast, read wherever the spell is as the
// cost is offered and paid. A colourless spell has no colour to share, so it
// gets no offer. Only where the printed mana cost could be paid (CR 118.9a), so
// a spell cast for another alternative cost is untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                "d901e17f-a5f4-49d9-8a6d-ab771949af24",
		Name:                    "Dream Halls",
		Completeness:            effects.CompletenessFull,
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{effects.DiscardAColorSharingCardForEverySpell()},
	})
}
