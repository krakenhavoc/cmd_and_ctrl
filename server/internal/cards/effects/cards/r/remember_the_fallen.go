package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Remember the Fallen — Sorcery {2}{W}:
//
//	"Choose one or both —
//	 • Return target creature card from your graveyard to your hand.
//	 • Return target artifact card from your graveyard to your hand."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ed3918d-509a-4fe6-a32d-5899df8f9046",
		Name:         "Remember the Fallen",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Return target creature card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target creature card in your graveyard", effects.YouOwn(), effects.Creature()),
				effects.S82OldB170ReturnModeTargetToHand),
			effects.ModeDoing("Return target artifact card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target artifact card in your graveyard", effects.YouOwn(), effects.Artifact()),
				effects.S82OldB170ReturnModeTargetToHand),
		),
	})
}
