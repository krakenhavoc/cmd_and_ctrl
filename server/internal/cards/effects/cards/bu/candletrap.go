package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Candletrap — Enchantment — Aura {W} (Midnight Hunt, #5497):
//
//	"Enchant creature
//	 Enchanted creature has defender.
//	 Prevent all combat damage that would be dealt by enchanted creature.
//	 Coven — {2}{W}, Sacrifice this Aura: Exile enchanted creature.
//	 Activate only if you control three or more creatures with different
//	 powers."
//
// Defender is a layer 6 grant to the host. The prevention is a CR 615.1a
// replacement that matches combat damage whose source is the host, read
// as the damage would be dealt. The activated ability is Path to
// Redemption's exile, gated on coven when it is activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "089788ce-d06c-4a85-915c-82e2cb5b0103",
		Name:         "Candletrap",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.GrantToAttached("defender")},
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a — "prevent"; CR 615.12 reads it
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return src != nil && ev.Kind == game.RepEventDamage && ev.IsCombatDamage &&
					ev.DamageAmount > 0 && src.AttachedTo.ID != uuid.Nil && ev.DamageSource == src.AttachedTo.ID
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Candletrap — prevent all combat damage dealt by enchanted creature",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:     "Coven — {2}{W}, Sacrifice this Aura: Exile enchanted creature. Activate only if you control three or more creatures with different powers.",
			Cost:      effects.Plus(effects.ManaCost("{2}{W}"), effects.SacrificeThis()),
			Condition: effects.TrekB48Coven,
			Purpose:   game.Purpose{Answers: game.AnswerRemove},
			Effect:    effects.ExileEnchantedCreature,
		}},
	})
}
