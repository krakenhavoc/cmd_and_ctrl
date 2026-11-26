package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savage Silhouette — Enchantment — Aura {2}{G} (Zendikar, #5583):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has "{1}{G}: Regenerate this
//	 creature.""
//
// Molting Snakeskin's shape: the creature, not the Aura, is the source of
// the granted regeneration row, so its controller pays and gets the
// shield, and both go with the Aura.
//
// No simplification.
func init() {
	const grant = "savage-silhouette/regenerate"
	const text = "{1}{G}: Regenerate this creature."
	effects.Register(effects.Spec{
		OracleID:     "bc243497-d95a-44e0-b7e3-c56b4ff7cd69",
		Name:         "Savage Silhouette",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{1}{G}"),
				Effect:  effects.RegenerateThisSource,
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{effects.PumpAttached(2, 2), effects.GrantAbilitiesToAttached(grant)},
	})
}
