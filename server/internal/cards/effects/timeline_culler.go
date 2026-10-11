package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Timeline Culler — Creature — Drix Warlock {B}{B}, 2/2:
//
//	"Haste
//	 You may cast this card from your graveyard using its warp ability.
//	 Warp—{B}, Pay 2 life. (You may cast this card from your hand or
//	 graveyard for its warp cost. If you do, exile this creature at the
//	 beginning of the next end step, then you may cast it from exile on a
//	 later turn.)"
//
// Warp is an alternative cost (CR 118.9, CR 702.185) with a life
// component, and it is offered from TWO zones, so the card declares two
// offers with two keys rather than one unbound one:
//
//   - "warp" is bound to the hand: {B} and 2 life.
//   - "warp-graveyard" is bound to the graveyard: the same price, same
//     delayed exile.
//
// Both carry Life 2, a cost validated before anything is paid: a player
// at 1 life is not offered either, and the life is never waived. The
// zone binding cuts both ways. The graveyard opens ONLY through the
// graveyard-bound offer, so a graveyard cast that claims nothing is
// refused for the printed {B}{B} (the card declares a priced graveyard
// path, and an ordinary-price cast from there is strictly better than
// printed), and the hand offer cannot be claimed from the graveyard.
//
// The exile and the "on a later turn" recast ride the shared warp clause
// (WarpExile): a creature warped from either zone is exiled at the
// beginning of the next end step and may be cast from exile, for its
// printed {B}{B}, on any later turn. The exile does not offer warp again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3c307122-bc18-41f1-87d5-ae6b3ca9b8b9",
		Name:            "Timeline Culler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		CastableZones:   []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{
			timelineCullerWarp("warp", game.ZoneHand),
			timelineCullerWarp("warp-graveyard", game.ZoneGraveyard),
		},
	})
}

// timelineCullerWarp is "Warp—{B}, Pay 2 life", bound to the zone it is
// cast from.
func timelineCullerWarp(key string, from game.ZoneKind) game.AlternativeCost {
	ac := Warp("{B}")
	ac.Key = key
	ac.Label = "Warp—{B}, Pay 2 life"
	ac.Life = 2
	ac.FromZone = from
	return ac
}
