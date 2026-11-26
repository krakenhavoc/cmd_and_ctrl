package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Singing Bell Strike — Enchantment — Aura {1}{U} (Khans of Tarkir):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step.
//	 Enchanted creature has "{6}: Untap this creature.""
//
// Bitter Chill's tap and untap lock, with a granted activated row (Presence of
// Gond's shape) that lets the creature's controller pay {6} to untap it: the
// untap happens only while it is still the permanent that was activated
// (CR 400.7).
//
// No simplification.
func init() {
	const grant = "singing-bell-strike/untap"
	const text = "{6}: Untap this creature."
	effects.Register(effects.Spec{
		OracleID:              "dad07f8c-1f24-4f56-b8a6-7c4b82e27e0f",
		Name:                  "Singing Bell Strike",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Singing Bell Strike — tap enchanted creature", effects.TapEnchantedCreatureOnEntry),
		},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{6}"),
				Effect:  effects.UntapThisPermanentEffect,
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
	})
}
