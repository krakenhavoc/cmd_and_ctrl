package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stray from History — Sorcery {2}{B} (Star Trek, #4973):
//
//	"Choose one —
//	 • Rewrite the Past — Return one or two target creature and/or
//	   Spacecraft cards from your graveyard to your hand.
//	 • Repair the Timeline — Target opponent reveals their hand. You
//	   choose a nonland card from it. Exile that card."
//
// The mode names are flavor words with no rules meaning (CR 207.2d). The
// first bullet's cards are re-checked one by one as it resolves
// (CR 608.2b): a card that has left the graveyard stays where it is and
// the other still returns. The second exiles the chosen card; it is not
// a discard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "adbbb968-5721-427c-a15a-320d49ab4c3e",
		Name:         "Stray from History",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Rewrite the Past — Return one or two target creature and/or Spacecraft cards from your graveyard to your hand.",
				effects.TargetCardInGraveyard("one or two target creature and/or Spacecraft cards from your graveyard",
					effects.YouOwn(), effects.Or(effects.Creature(), effects.Subtype("Spacecraft"))).WithCount(1, 2),
				effects.ReturnTheModesTargetsToHand),
			effects.ModeDoing("Repair the Timeline — Target opponent reveals their hand. You choose a nonland card from it. Exile that card.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.ChooseFromRevealedHand{Player: t.ID, Filter: effects.Nonland(), Label: "nonland card", Exile: true}.Apply(ctx)
				}),
		),
	})
}
