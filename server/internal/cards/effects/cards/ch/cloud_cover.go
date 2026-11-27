package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloud Cover — Enchantment {2}{W}{U} (Planeshift, S82 OLD b644, #5615):
//
//	"Whenever another permanent you control becomes the target of a spell or
//	 ability an opponent controls, you may return that permanent to its owner's
//	 hand."
//
// Mila's targeting watch with "another": the event is emitted once per target
// slot (CR 115.3), the permanent must be on the battlefield under your control, not
// this enchantment, and the spell or ability must be an opponent's. The "you may"
// is asked as the trigger resolves (CR 603.5); returning the permanent in response
// leaves the spell or ability without that target.
//
// No simplification.
func init() {
	const label = "Cloud Cover — return that permanent to its owner's hand"
	effects.Register(effects.Spec{
		OracleID:     "a4cd8dd1-adb9-4c4c-bda9-6f6b6c167702",
		Name:         "Cloud Cover",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventBecomesTarget, effects.OldB644AnotherPermanentYouControlTargeted, label, effects.OldB644BounceTriggeringPermanent), label+"?"),
		},
	})
}
