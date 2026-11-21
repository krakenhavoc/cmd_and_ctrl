package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lotus Ring — Artifact — Equipment {3} (The Big Score, #4468):
//
//	"Indestructible
//	 Equipped creature gets +3/+3 and has vigilance and '{T}, Sacrifice
//	 this creature: Add three mana of any one color.'
//	 Equip {3}"
//
// The grant is Alchemist's Talent's shape (an ADR 0093 bundle with a mana
// ability whose cost sacrifices its source), here the equipped creature.
//
// No simplification.
func init() {
	const grant = "lotus-ring/three-mana-of-one-color"
	effects.Register(effects.Spec{
		OracleID:        "ca4d3d27-ea3a-44ac-96e9-b52b563129f8",
		Name:            "Lotus Ring",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Grants: []effects.AbilityGrant{{
			Key:  grant,
			Text: "{T}, Sacrifice this creature: Add three mana of any one color.",
			Mana: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Tap: true, Sacrifice: true},
				Produced: effects.OneColorOfAmount(3),
				Label:    "Add three mana of any one color",
			}},
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.GrantToAttached("vigilance"),
			effects.GrantAbilitiesToAttached(grant),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
