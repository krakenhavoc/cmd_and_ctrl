package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trailblazer's Torch — Artifact — Equipment {4}:
//
//	"When this Equipment enters, you take the initiative.
//	 Whenever equipped creature becomes blocked, it deals 2 damage to
//	 each creature blocking it.
//	 Equip {1}"
//
// The ETB is TakeTheInitiative (ADR 0203, #4096). "Becomes blocked" is
// once per blocked attacker however many creatures block it
// (EventBecomesBlocked, CR 509.3c), and "each creature blocking it" is
// read as the trigger resolves, so a blocker removed in response is not
// dealt damage. "It" is the equipped creature: the damage is that
// creature's, dealt as one instruction to every blocker (one damage
// instance, CR 615.5). Equip is the shared EquipAbility.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1cdac498-3c94-48d6-b94c-1d98697370ea",
		Name:         "Trailblazer's Torch",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouTakeTheInitiative("Trailblazer's Torch"),
			On(game.EventBecomesBlocked, trailblazersTorchEquippedBlocked,
				"Trailblazer's Torch — the equipped creature deals 2 damage to each creature blocking it", trailblazersTorchDamage),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}

// trailblazersTorchEquippedBlocked: the creature this Equipment is
// attached to became blocked.
func trailblazersTorchEquippedBlocked(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && ev.Kind == game.EventBecomesBlocked && ev.CardID != source.InstanceID && source.IsAttachedTo(ev.CardID)
}

// trailblazersTorchDamage has the blocked creature deal 2 damage to
// each creature blocking it now.
func trailblazersTorchDamage(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	attacker := item.Trigger.Event.CardID
	blockers := blockersOf(g, attacker)
	if len(blockers) == 0 {
		return nil
	}
	return g.DealDamageEachKeyedForEffect(attacker, blockers, 2, game.KeyedTail{})
}
