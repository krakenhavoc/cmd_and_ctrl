package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Season of Renewal — Instant {2}{G} (Kamigawa: Neon Dynasty, #5493):
//
//	"Choose one or both —
//	 • Return target creature card from your graveyard to your hand.
//	 • Return target enchantment card from your graveyard to your hand."
//
// A target that left the graveyard by resolution does nothing (CR 608.2b);
// the other mode still resolves.
//
// No simplification.
func init() {
	back := func(label string, pred effects.CardPredicate, text string) game.ModeOption {
		return effects.ModeDoing(text,
			effects.TargetCardInGraveyard(label, effects.YouOwn(), pred),
			effects.ReturnTheModesGraveyardTargetToHand)
	}
	effects.Register(effects.Spec{
		OracleID:     "410ce48e-d907-4bc5-ad90-2067161a0e04",
		Name:         "Season of Renewal",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOneOrMore(
			back("target creature card in your graveyard", effects.Creature(),
				"Return target creature card from your graveyard to your hand."),
			back("target enchantment card in your graveyard", effects.Enchantment(),
				"Return target enchantment card from your graveyard to your hand."),
		),
	})
}
