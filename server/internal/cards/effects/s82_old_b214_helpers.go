package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b214_helpers.go — S82 OLD batch 214 (Dominaria, #5530).

// OldB214DestroyAttachedEquipmentBody is Corrosive Ooze's delayed trigger at
// end of combat: destroy every Equipment attached, as the trigger resolves, to
// the creature the item names.
var OldB214DestroyAttachedEquipmentBody = game.SimpleDelayedBody("s82-old-b214/destroy-attached-equipment", oldB214DestroyAttachedEquipment)

func oldB214DestroyAttachedEquipment(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil {
			continue
		}
		for _, id := range g.AttachmentsOf(t.ID) {
			c, ok := g.LookupCardForEffect(id)
			if !ok || !c.HasSubtype("Equipment") {
				continue
			}
			if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// OldB214PreventOfDamageToEquipped is "If a source would deal damage to
// equipped creature, prevent N of that damage" (Shield of the Realm, CR 615):
// damage of N or less is prevented entirely, larger damage loses N.
func OldB214PreventOfDamageToEquipped(n int, label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true, // CR 615.1a
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return src != nil && ev.Kind == game.RepEventDamage && ev.DamageAmount > 0 && src.IsAttachedTo(ev.DamageTarget)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			if ev.DamageAmount <= n {
				ev.Cancel()
				return nil
			}
			ev.DamageAmount -= n
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      label,
		RowLabel:   "Prevent 2 of the damage that would be dealt to the equipped creature.",
	}
}
