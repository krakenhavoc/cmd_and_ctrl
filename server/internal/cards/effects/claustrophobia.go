package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Claustrophobia —
//
// "Enchant creature When this Aura enters, tap enchanted creature. Enchanted
// creature doesn't untap during its controller's untap step."
//
// The entry trigger reads the current attachment when it resolves. If the Aura
// has left, its last known attachment is not yet available at resolution, so
// that case remains a declared caveat.
func init() {
	Register(Spec{
		OracleID:              "62d8c8c8-bc24-42f2-9e2e-9efd08e47bb1",
		Name:                  "Claustrophobia",
		Completeness:          CompletenessCaveats,
		Caveats:               []string{"If Claustrophobia leaves the battlefield before its enter ability resolves, it does not tap the creature."},
		Targets:               EnchantCreature(),
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{WhenThisEnters("Claustrophobia — tap enchanted creature", func(g *game.Game, item *game.StackItem) error {
			c, ok := g.LookupCardForEffect(item.SourceCardID)
			if !ok || c.AttachedTo.Kind != game.TargetCard {
				return nil
			}
			return TapTarget{Target: c.AttachedTo.ID}.Apply(NewContext(g, item))
		})},
	})
}
