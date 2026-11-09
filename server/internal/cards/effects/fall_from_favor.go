package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fall from Favor — Enchantment — Aura {2}{U} (Foundations Commander):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature and you become the
//	 monarch.
//	 Enchanted creature doesn't untap during its controller's untap step
//	 unless that player is the monarch."
//
// Claustrophobia's tap and Dramatic Accusation's untap lock, with the
// printed exception: the restriction is asked at each of the creature's
// controller's untap steps (CR 502.3) and lifts while that player is the
// monarch, so the step after they take the crown untaps it. The tap comes
// first, then the crown (CR 608.2c). The restriction is read live from the
// Aura on the battlefield, so it ends when the Aura leaves.
//
// No simplification.
func init() {
	lock := enchantedDoesntUntap()
	attached := lock.Restricts
	lock.Label = "enchanted creature doesn't untap unless its controller is the monarch"
	lock.Restricts = func(target *game.Card, g *game.Game, source *game.Card) bool {
		return attached(target, g, source) && g.Monarch != target.Controller
	}
	Register(Spec{
		OracleID:              "e89e1fae-bac3-4cc0-a233-24b9e08a5290",
		Name:                  "Fall from Favor",
		Completeness:          CompletenessFull,
		Targets:               EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{lock},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fall from Favor — tap enchanted creature and you become the monarch",
				Do(Instruction(func(ctx *Context) error {
					return tapEnchantedCreatureOnEntry(ctx.Game, ctx.Item)
				}), BecomeTheMonarch{})),
		},
	})
}
