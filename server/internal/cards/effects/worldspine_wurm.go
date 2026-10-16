package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Worldspine Wurm — Creature — Wurm {8}{G}{G}{G}, 15/15:
//
//	"Trample
//	 When this creature dies, create three 5/5 green Wurm creature
//	 tokens with trample.
//	 When Worldspine Wurm is put into a graveyard from anywhere, shuffle
//	 it into its owner's library."
//
// When it dies both abilities trigger, and the tokens come whichever
// resolves first, since the dies
// trigger looks back in time (CR 603.10a) and does not need the card.
// Put into a graveyard from anywhere else, only the shuffle fires. The shuffle is
// the graveyard trigger (#2038,
// WhenThisIsPutIntoAGraveyardFromAnywhereShuffleItIn).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "254f87cc-e0c2-484d-b5e8-1ea8c366990c",
		Name:            "Worldspine Wurm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Worldspine Wurm — create three 5/5 green Wurm tokens with trample",
				Do(CreateToken{Template: TokenCard("5/5 green Wurm with trample"), N: 3})),
			WhenThisIsPutIntoAGraveyardFromAnywhereShuffleItIn("Worldspine Wurm — shuffle it into its owner's library"),
		},
	})
}
