package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Claustrophobia —
//
// "Enchant creature When this Aura enters, tap enchanted creature. Enchanted
// creature doesn't untap during its controller's untap step."
//
// The entry trigger reads the Aura's attachment when it resolves: the
// creature it enchants now, and — if the Aura has left the battlefield in
// response — the creature it last enchanted (CR 608.2h, #1379). That
// creature is still tapped; only the "doesn't untap" half ends with the
// Aura.
func init() {
	Register(Spec{
		OracleID:              "62d8c8c8-bc24-42f2-9e2e-9efd08e47bb1",
		Name:                  "Claustrophobia",
		Completeness:          CompletenessFull,
		Targets:               EnchantCreature(),
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{WhenThisEnters("Claustrophobia — tap enchanted creature", func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			aura, ok := ctx.TriggeringPermanent()
			if !ok || aura.AttachedTo.Kind != game.TargetCard {
				return nil
			}
			return TapTarget{Target: aura.AttachedTo.ID}.Apply(ctx)
		})},
	})
}
