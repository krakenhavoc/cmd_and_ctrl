package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "ca4d3d27-ea3a-44ac-96e9-b52b563129f8",
		Name:            "Lotus Ring",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Grants: []AbilityGrant{{
			Key:  grant,
			Text: "{T}, Sacrifice this creature: Add three mana of any one color.",
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true, Sacrifice: true},
				Produced: OneColorOfAmount(3),
				Label:    "Add three mana of any one color",
			}},
		}},
		Static: []game.StaticAbility{
			PumpAttached(3, 3),
			GrantToAttached("vigilance"),
			GrantAbilitiesToAttached(grant),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
