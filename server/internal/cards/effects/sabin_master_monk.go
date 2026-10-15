package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sabin, Master Monk — Legendary Creature — Human Noble Monk {4}{R}, 4/3:
//
//	"Double strike
//	 Blitz—{2}{R}{R}, Discard a card. (If you cast this spell for its
//	 blitz cost, it gains haste and "When this creature dies, draw a
//	 card." Sacrifice it at the beginning of the next end step.)
//	 You may cast this card from your graveyard using its blitz ability."
//
// Blitz (CR 702.152, #3772) offered from TWO zones, so the card declares
// two offers with two keys, the shape Timeline Culler's warp uses:
// "blitz" bound to the hand and "blitz-graveyard" bound to the graveyard,
// both {2}{R}{R} plus a discard. The graveyard opens only through the
// graveyard offer, so Sabin cannot be cast from there for {4}{R}.
//
// A blitzed Sabin that is sacrificed at the end step dies, draws a card
// and is back in the graveyard, ready to be blitzed again next turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "7234bbf8-b962-4515-bedf-68d111c6137e",
		Name:             "Sabin, Master Monk",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"double strike"},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: BlitzFromHandOrGraveyard(BlitzDiscard("{2}{R}{R}")),
	})
}
