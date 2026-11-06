package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nervous Gardener — Creature — Dryad {1}{G}, 2/2 (Murders at Karlov
// Manor, #4469):
//
//	"Disguise {G} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)
//	 When this creature is turned face up, search your library for a
//	 land card with a basic land type, reveal it, put it into your hand,
//	 then shuffle."
//
// Boseiju, Who Endures's "land card with a basic land type"
// (hasABasicLandType: the land type, so a dual with basic types counts
// and Wastes does not), fetched to the hand and revealed. A search for a
// card with a stated quality may find nothing (CR 701.23b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "15dca6a7-8367-4b86-a86c-e44f10fb3c53",
		Name:             "Nervous Gardener",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Disguise("{G}")},
		Triggered: []game.TriggeredAbility{
			WhenThisIsTurnedFaceUp("Nervous Gardener — search your library for a land card with a basic land type",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return SearchLibrary{
						Player:    ctx.Controller(),
						Predicate: hasABasicLandType,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Nervous Gardener — search for a land card with a basic land type",
					}.Apply(ctx)
				}),
		},
	})
}
