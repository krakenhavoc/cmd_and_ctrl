package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	back := func(label string, pred CardPredicate, text string) game.ModeOption {
		return ModeDoing(text,
			TargetCardInGraveyard(label, YouOwn(), pred),
			ReturnTheModesGraveyardTargetToHand)
	}
	Register(Spec{
		OracleID:     "410ce48e-d907-4bc5-ad90-2067161a0e04",
		Name:         "Season of Renewal",
		Completeness: CompletenessFull,
		Modes: ChooseOneOrMore(
			back("target creature card in your graveyard", Creature(),
				"Return target creature card from your graveyard to your hand."),
			back("target enchantment card in your graveyard", Enchantment(),
				"Return target enchantment card from your graveyard to your hand."),
		),
	})
}
