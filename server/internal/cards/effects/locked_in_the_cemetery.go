package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Locked in the Cemetery — Enchantment — Aura {1}{U} (Innistrad: Midnight
// Hunt, S82):
//
//	"Enchant creature
//	 When this Aura enters, if there are five or more cards in your
//	 graveyard, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Claustrophobia's lock with the tap behind an intervening "if" (CR 603.4):
// the graveyard count is checked as the Aura enters and again as the
// trigger resolves, and a trigger that fails either does nothing. The
// "doesn't untap" half applies either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "035f0aae-9c23-4b22-8c87-0d66b84b7962",
		Name:                  "Locked in the Cemetery",
		Completeness:          CompletenessFull,
		Targets:               EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{Name: "there are five or more cards in your graveyard", Test: func(ctx *Context) bool {
				p := ctx.PlayerByID(ctx.Controller())
				return p != nil && p.Graveyard != nil && p.Graveyard.Size() >= 5
			}}, WhenThisEnters("Locked in the Cemetery — tap enchanted creature", tapEnchantedCreatureOnEntry)),
		},
	})
}
