package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Entomber Exarch — Creature — Phyrexian Cleric {2}{B}{B}, 2/2:
//
//	"When this creature enters, choose one —
//	 • Return target creature card from your graveyard to your hand.
//	 • Target opponent reveals their hand. You choose a noncreature
//	   card from it. That player discards that card."
//
// A modal enters trigger (CR 700.2, Charming Prince's shape): the mode
// and its target are chosen as the trigger goes on the stack, and each
// bullet's clause rides its own ModeOption. The second bullet is the
// revealed-hand pick (ADR 0116) filtered by Noncreature(), so a land is
// a legal pick, as printed. The chooser is the trigger's controller
// (CR 113.8). A hand with no noncreature card is revealed and nothing
// is discarded (CR 609.3).
//
// No simplification.
func init() {
	exarch := WhenThisEnters("Entomber Exarch — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	exarch.Modes = ChooseOne(
		ModeDoing("Return target creature card from your graveyard to your hand.",
			TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
			ReturnTheModesGraveyardTargetToHand),
		ModeDoing("Target opponent reveals their hand. You choose a noncreature card from it. That player discards that card.",
			TargetPlayer("target opponent", Opponent()),
			ModeTargetRevealsYouChooseDiscard(Noncreature(), "noncreature card")),
	)
	Register(Spec{
		OracleID:     "e820296a-81b0-401f-959c-7aed8abefce1",
		Name:         "Entomber Exarch",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{exarch},
	})
}
