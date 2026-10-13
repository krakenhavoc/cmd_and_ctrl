package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// ThatMuchDamage makes a "whenever ~ is dealt damage, … that much"
// ability read the TOTAL of the damage dealt to the object at the same
// time (CR 120.3, CR 510.2, CR 603.2c), not each source's share: the
// ability waits for its batch to settle, fires once per damaged object
// and damage instance, and `item.Trigger.Event.Amount` is the sum. Set
// it as `AtBatchEnd:` on the ability; it needs a non-empty Key. See
// game/batch_total.go (ADR 0108 amendment, 2026-10-13).
func ThatMuchDamage() *game.BatchEndCondition {
	return &game.BatchEndCondition{Total: game.BatchTotalDamage}
}

// ThatManyEvents is the counting twin for "whenever you [do X] one or
// more times … that many" (Cryptcaller Chariot, Magmakin Artillerist,
// Ingenious Artillerist): one trigger per batch, with
// `item.Trigger.Event.Amount` the number of events the ability's
// AppliesTo accepted.
func ThatManyEvents() *game.BatchEndCondition {
	return &game.BatchEndCondition{Total: game.BatchTotalEvents}
}

// withAtBatchEnd returns t with its AtBatchEnd condition set, for the
// constructors (On, OnAny) that build the whole ability from a label.
func withAtBatchEnd(t game.TriggeredAbility, c *game.BatchEndCondition) game.TriggeredAbility {
	t.AtBatchEnd = c
	return t
}

// settledEventAmount is the total a ThatManyEvents / ThatMuchDamage
// ability settled with: the triggering event's Amount, which the
// settle pass rewrote (game/batch_total.go). At least 1 for a trigger
// that somehow reaches resolution without one.
func settledEventAmount(item *game.StackItem) int {
	if item.Trigger == nil || item.Trigger.Event.Amount < 1 {
		return 1
	}
	return item.Trigger.Event.Amount
}
