package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// ringingStrikeMasteryGrant is the bundle the Aura gives the enchanted
// creature. An on-disk identity: never renamed.
const ringingStrikeMasteryGrant = "ringing-strike-mastery/untap"

// Ringing Strike Mastery — Enchantment — Aura {U} (Tarkir: Dragonstorm,
// #4463):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step.
//	 Enchanted creature has "{5}: Untap this creature.""
//
// Claustrophobia with a way out: the tap on entry and the untap-step rule
// are Claustrophobia's, and the granted ability is a layer-6 bundle on the
// enchanted creature (ADR 0093). Its controller activates it, and "this
// creature" is the enchanted creature, untapped if it is still on the
// battlefield as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "334ffdc4-9fe6-4a7f-bf5d-bcd88adab715",
		Name:                  "Ringing Strike Mastery",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Grants: []effects.AbilityGrant{{
			Key: ringingStrikeMasteryGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{5}: Untap this creature.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{5}"),
				Effect:  effects.B31UntapSelf,
			}},
			Text: "{5}: Untap this creature.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(ringingStrikeMasteryGrant)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ringing Strike Mastery — tap enchanted creature", effects.TapEnchantedCreatureOnEntry),
		},
	})
}
