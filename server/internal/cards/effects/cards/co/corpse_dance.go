package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Corpse Dance — {2}{B} Instant (Tempest, S82 OLD batch 769, #5628):
//
//	"Buyback {2} (You may pay an additional {2} as you cast this spell. If you
//	 do, put this card into your hand as it resolves.)
//	 Return the top creature card of your graveyard to the battlefield. That
//	 creature gains haste until end of turn. Exile it at the beginning of the
//	 next end step."
//
// Mind Games' buyback (CR 702.27a) over Shallow Grave's body
// (S82OldB769ReturnTopCreatureHasteExile): the topmost CREATURE card of the
// graveyard, found as the spell resolves, so a card milled in response is the one
// that returns. With the buyback paid the spell returns to its owner's hand as it
// resolves, so a Corpse Dance cast with nothing to return still comes back.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "59b02ca7-e7bd-4820-ac32-87ea1a8c64c0",
		Name:          "Corpse Dance",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Buyback("{2}")},
		OnResolve:     effects.S82OldB769ReturnTopCreatureHasteExile("Corpse Dance"),
	})
}
