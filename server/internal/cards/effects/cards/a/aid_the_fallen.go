package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aid the Fallen — Sorcery {1}{B} (War of the Spark, #5520):
//
//	"Choose one or both —
//	 • Return target creature card from your graveyard to your hand.
//	 • Return target planeswalker card from your graveyard to your hand."
//
// Borrowed Grace's "one or both" (ChooseN 1..2), each bullet with its own
// target in your graveyard (CR 700.2d). A bullet whose target has left the
// graveyard does nothing; the other still happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "acb0084d-3b09-4d2a-a4c1-848c92588845",
		Name:         "Aid the Fallen",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Return target creature card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target creature card in your graveyard", effects.YouOwn(), effects.Creature()),
				effects.S82OldB170ReturnModeTargetToHand),
			effects.ModeDoing("Return target planeswalker card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target planeswalker card in your graveyard", effects.YouOwn(), effects.Planeswalker()),
				effects.S82OldB170ReturnModeTargetToHand),
		),
	})
}
