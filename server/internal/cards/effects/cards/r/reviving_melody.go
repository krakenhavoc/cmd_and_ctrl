package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reviving Melody — Sorcery {2}{G} (Journey into Nyx, #5558):
//
//	"Choose one or both —
//	 • Return target creature card from your graveyard to your hand.
//	 • Return target enchantment card from your graveyard to your hand."
//
// Season of Renewal's "choose one or both": each bullet returns its own
// graveyard target, in printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4834e6f2-1544-4e58-9092-ecc136cc0dee",
		Name:         "Reviving Melody",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOneOrMore(
			effects.ModeDoing("Return target creature card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target creature card in your graveyard", effects.YouOwn(), effects.Creature()),
				effects.ReturnTheModesGraveyardTargetToHand),
			effects.ModeDoing("Return target enchantment card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target enchantment card in your graveyard", effects.YouOwn(), effects.Enchantment()),
				effects.ReturnTheModesGraveyardTargetToHand),
		),
	})
}
