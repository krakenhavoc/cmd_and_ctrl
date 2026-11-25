package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b077_helpers.go — the Rock Equipment token Toggo, Goblin
// Weaponsmith makes (Commander Legends, S82 OLD batch 77, #5507). It lives
// in the core because the token catalog lists it.

const s82OldB077RockGrant = "rock-equipment/deal-2"

// RockEquipmentToken is the Rock as PRINTED: a colorless Equipment artifact
// token named Rock, "Equipped creature has '{1}, {T}, Sacrifice Rock: This
// creature deals 2 damage to any target.'", equip {1}.
func RockEquipmentToken() game.Card { return TokenFromCatalog(printedS82OldB077RockToken) }

// printedS82OldB077RockToken is the catalog's entry for the Rock. Its slug
// and grant key are on-disk identities.
func printedS82OldB077RockToken() TokenTemplate {
	const text = "{1}, {T}, Sacrifice Rock: This creature deals 2 damage to any target."
	return TokenTemplate{
		Slug: "rock-equipment",
		Card: game.Card{
			Name:     "Rock",
			TypeLine: "Token Artifact — Equipment",
		},
		Text: "Equipped creature has \"" + text + "\"\nEquip {1}",
		Grants: []AbilityGrant{{
			Key: s82OldB077RockGrant,
			Activated: []ActivatedAbility{{
				Label:   text,
				Purpose: game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove},
				Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeTheGrantor()),
				Targets: TargetAny(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if len(item.Targets) == 0 {
						return nil
					}
					// "This creature" is the equipped creature, the ability's source.
					return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: 2}.Apply(ctx)
				},
			}},
			Text: text,
		}},
		Static:    []game.StaticAbility{GrantAbilitiesToAttached(s82OldB077RockGrant)},
		Activated: ActivatedShapes([]ActivatedAbility{EquipAbility("{1}")}),
	}
}
