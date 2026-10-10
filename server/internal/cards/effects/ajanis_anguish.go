package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ajani's Anguish — Enchantment {X}{R} (Reality Fracture):
//
//	"When this enchantment enters, it deals X damage to any target.
//	 Creatures you control have trample."
//
// X is the value announced for the spell (CR 107.3m), read off the
// permanent as the trigger is built and carried on the item, as
// Arboreal Alliance does. A copy or a cheat onto the battlefield has
// X = 0 and deals no damage. The damage is dealt by the enchantment, so
// it is gone from the battlefield only if it left before the trigger
// resolved (last known information still names it as the source).
//
// No simplification.
func init() {
	const label = "Ajani's Anguish — it deals X damage to any target"
	Register(Spec{
		OracleID:     "4a37f4a0-2869-4f81-b51b-4c3e7e0e0671",
		Name:         "Ajani's Anguish",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{YoursOnly: true}, "trample"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Targets:   TargetAny(),
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
		}},
	})
}
