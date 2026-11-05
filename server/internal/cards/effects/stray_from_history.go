package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "adbbb968-5721-427c-a15a-320d49ab4c3e",
		Name:         "Stray from History",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Rewrite the Past — Return one or two target creature and/or Spacecraft cards from your graveyard to your hand.",
				TargetCardInGraveyard("one or two target creature and/or Spacecraft cards from your graveyard",
					YouOwn(), Or(Creature(), Subtype("Spacecraft"))).WithCount(1, 2),
				ReturnTheModesTargetsToHand),
			ModeDoing("Repair the Timeline — Target opponent reveals their hand. You choose a nonland card from it. Exile that card.",
				TargetPlayer("target opponent", Opponent()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return ChooseFromRevealedHand{Player: t.ID, Filter: Nonland(), Label: "nonland card", Exile: true}.Apply(ctx)
				}),
		),
	})
}
