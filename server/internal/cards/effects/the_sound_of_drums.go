package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Sound of Drums — Enchantment — Aura {2}{R} (Doctor Who):
//
//	"Enchant creature
//	 Enchanted creature is goaded.
//	 If enchanted creature would deal combat damage to a permanent or
//	 player, it deals double that damage instead.
//	 {2}{R}: Return this card from your graveyard to your hand."
//
// The goad is GoadAttached (a static one, CR 701.15b, by the Aura's
// controller). The doubling is Dictate of the Twin Gods' replacement
// narrowed to combat damage dealt by the creature this Aura is attached to
// (CR 614.1, 120.3). The last ability works only from the graveyard
// (Abzan Devotee's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0afcc646-cb34-4c0e-b3f5-4115ad4faa8e",
		Name:         "The Sound of Drums",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{GoadAttached()},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventDamage && ev.DamageAmount > 0 && ev.IsCombatDamage && src.IsAttachedTo(ev.Source)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			RowLabel: "If enchanted creature would deal combat damage to a permanent or player, it deals double that damage instead.",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{2}{R}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  returnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
