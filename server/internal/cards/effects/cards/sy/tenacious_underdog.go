package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tenacious Underdog — Creature — Human Warrior {1}{B}, 3/2:
//
//	"Blitz—{2}{B}{B}, Pay 2 life. (If you cast this spell for its blitz
//	 cost, it gains haste and "When this creature dies, draw a card."
//	 Sacrifice it at the beginning of the next end step.)
//	 You may cast this card from your graveyard using its blitz ability."
//
// Blitz (CR 702.152, #3772) with a life component, offered from the hand
// and from the graveyard (BlitzFromHandOrGraveyard): the graveyard opens
// only through the graveyard-bound offer, so the Underdog cannot be cast
// from there for {1}{B}. The 2 life is validated before anything is paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "598e3ed2-659f-4400-b41d-3aee8068fba6",
		Name:             "Tenacious Underdog",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: effects.BlitzFromHandOrGraveyard(effects.BlitzPayLife("{2}{B}{B}", 2)),
	})
}
