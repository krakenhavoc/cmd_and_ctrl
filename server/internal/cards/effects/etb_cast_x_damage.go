package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// etbDealsCastXDamage is "When this enters, it deals X damage to <target>"
// for a permanent cast with X (Ajani's Anguish, Morlun, Devourer of Spiders).
// X is the value announced for the spell (CR 107.3m), read off the permanent
// as the trigger is built and carried on the item, so it is exact even if the
// source has left by the time the trigger resolves. A copy or a cheat onto
// the battlefield has X = 0 and deals no damage. The damage is the source's
// (last known information if it left), so lifelink and Torbran apply.
func etbDealsCastXDamage(label string, targets *game.TargetSpec) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventETB},
		AppliesTo: Self,
		Targets:   targets,
		Key:       label,
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Amount = source.CastX()
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			x := item.Params.Amount
			if x <= 0 {
				return nil
			}
			for _, t := range ctx.LegalTargets() {
				return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: x}.Apply(ctx)
			}
			return nil
		},
	}
}
