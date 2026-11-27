package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grim Discovery — Sorcery {1}{B} (Zendikar, #5583):
//
//	"Choose one or both —
//	 • Return target creature card from your graveyard to your hand.
//	 • Return target land card from your graveyard to your hand."
//
// Saheeli's Artistry's "one or both" with Disentomb's return on each
// bullet's own target; a bullet whose card has left the graveyard does
// nothing and the other still happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be78d7ca-b904-452f-b9eb-34792e588986",
		Name:         "Grim Discovery",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Return target creature card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()), effects.NeoB08ReturnTheModesTargetToHand),
			effects.ModeDoing("Return target land card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target land card from your graveyard", effects.YouOwn(), effects.Land()), effects.NeoB08ReturnTheModesTargetToHand),
		),
	})
}
