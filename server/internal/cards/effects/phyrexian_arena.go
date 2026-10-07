package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Arena — Enchantment for {1}{B}{B}:
//
//	"At the beginning of your upkeep, you draw a card and lose 1
//	 life."
//
// S19 sub-PR 5: the first "your upkeep" trigger on the auto-fire
// pipeline. EventBeginUpkeep carries the active player in Actor;
// AppliesTo gates on Actor == Controller so the Arena only fires on
// its own controller's upkeep (not every player's). Mandatory: the
// trigger goes on the stack at upkeep and the life loss + draw
// happen when it resolves, in the printed order: the card is drawn
// first, then the life is lost (a draw trigger sees the pre-loss
// life total, and a controller at 1 life still draws before the
// state-based action ends them).
func init() {
	Register(Spec{
		OracleID:     "ee579a32-a048-4335-b966-231ba731cdea",
		Name:         "Phyrexian Arena",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Phyrexian Arena — draw a card, lose 1 life", b36DrawAndLoseOne),
		},
	})
}
