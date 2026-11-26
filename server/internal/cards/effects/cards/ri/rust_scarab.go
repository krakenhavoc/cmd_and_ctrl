package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rust Scarab — Creature — Insect {4}{G}, 4/5 (Gatecrash, #5565):
//
//	"Whenever this creature becomes blocked, you may destroy target artifact
//	 or enchantment defending player controls."
//
// The trigger is Battle-Scarred Goblin's (EventBecomesBlocked on itself,
// once however many creatures block it, CR 509.3c). The target clause is
// built per trigger from the player the Scarab is attacking (CR 802.2a:
// Commander's multiple defending players), as Havoc Eater's creature clause
// is; the "may" is asked as the ability resolves, the target having been
// chosen as it went on the stack (CR 603.3d), and the target is checked again
// then (CR 608.2b).
//
// No simplification.
func init() {
	t := effects.On(game.EventBecomesBlocked, effects.Self,
		"Rust Scarab — you may destroy target artifact or enchantment defending player controls",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			for _, tg := range ctx.LegalTargets() {
				if tg.Kind == game.TargetCard {
					return effects.DestroyTarget{Target: tg.ID}.Apply(ctx)
				}
			}
			return nil
		})
	t.TargetsFrom = effects.TargetArtifactOrEnchantmentDefendingPlayerControls
	t = effects.Optional(t, "Rust Scarab — destroy target artifact or enchantment defending player controls?")
	effects.Register(effects.Spec{
		OracleID:     "c3bd1d38-e73f-4e30-b22a-dca0a0e4d701",
		Name:         "Rust Scarab",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
