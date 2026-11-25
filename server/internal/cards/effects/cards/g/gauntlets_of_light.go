package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// gauntletsOfLightGrant is the bundle the Aura gives the enchanted creature. An
// on-disk identity: never renamed.
const gauntletsOfLightGrant = "gauntlets-of-light/untap"

// Gauntlets of Light — Enchantment — Aura {2}{W} (Core Set 2020, #5518):
//
//	"Enchant creature
//	 Enchanted creature gets +0/+2 and assigns combat damage equal to its
//	 toughness rather than its power.
//	 Enchanted creature has "{2}{W}: Untap this creature.""
//
// A layer 7c pump, Felothar the Steadfast's DamageByToughnessAlways on the
// host (CR 510.1c; toughness whether or not it is the greater), and a layer 6
// bundle (ADR 0093) giving the host an untap ability its controller pays for.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5741c85-cd09-4ccb-ab39-3f3a62f98939",
		Name:         "Gauntlets of Light",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: gauntletsOfLightGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{2}{W}: Untap this creature.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{2}{W}"),
				Effect:  effects.B31UntapSelf,
			}},
			Text: "{2}{W}: Untap this creature.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(0, 2),
			effects.OldB155DamageByToughness(),
			effects.GrantAbilitiesToAttached(gauntletsOfLightGrant),
		},
	})
}
