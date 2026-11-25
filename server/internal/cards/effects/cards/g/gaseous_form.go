package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gaseous Form — Enchantment — Aura {2}{U} (Eternal Masters, S82 OLD, #5542):
//
//	"Enchant creature
//	 Prevent all combat damage that would be dealt to and dealt by enchanted
//	 creature."
//
// Candletrap's CR 615.1a prevention of combat damage dealt by the host, widened
// to the same damage dealt to it: one replacement that matches a combat damage
// event whose source or recipient is the creature this Aura is attached to,
// read as the damage would be dealt. Noncombat damage is untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "896af9a9-12a2-4dd6-957c-e52150b5f3a2",
		Name:         "Gaseous Form",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a — "prevent"; CR 615.12 reads it
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				if src == nil || ev.Kind != game.RepEventDamage || !ev.IsCombatDamage || ev.DamageAmount <= 0 || src.AttachedTo.ID == uuid.Nil {
					return false
				}
				return ev.DamageSource == src.AttachedTo.ID || ev.DamageTarget == src.AttachedTo.ID
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Gaseous Form — prevent all combat damage dealt to and dealt by enchanted creature",
		}},
	})
}
