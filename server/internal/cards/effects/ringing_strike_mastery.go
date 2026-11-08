package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:              "334ffdc4-9fe6-4a7f-bf5d-bcd88adab715",
		Name:                  "Ringing Strike Mastery",
		Completeness:          CompletenessFull,
		Targets:               EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Grants: []AbilityGrant{{
			Key: ringingStrikeMasteryGrant,
			Activated: []ActivatedAbility{{
				Label:   "{5}: Untap this creature.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    ManaCost("{5}"),
				Effect:  b31UntapSelf,
			}},
			Text: "{5}: Untap this creature.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(ringingStrikeMasteryGrant)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ringing Strike Mastery — tap enchanted creature", tapEnchantedCreatureOnEntry),
		},
	})
}
