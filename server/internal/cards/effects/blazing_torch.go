package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blazing Torch — Artifact — Equipment {1} (Innistrad Remastered, #5948):
//
//	"Equipped creature can't be blocked by Vampires or Zombies.
//	 Equipped creature has "{T}, Sacrifice Blazing Torch: Blazing Torch deals 2
//	 damage to any target."
//	 Equip {1} ({1}: Attach to target creature you control. Equip only as a
//	 sorcery.)"
//
// The first sentence is a block rule on the attached creature, the Vampire and
// Zombie subtypes read as words of the Torch's text. The second is Ninja's
// Kunai's grant: the creature taps and the Torch is sacrificed (as the grantor)
// to pay, and the damage is dealt by the Torch, as it last existed
// (CR 608.2h), to the one target.
//
// No simplification.
const blazingTorchGrant = "blazing-torch/deal-2"

func init() {
	Register(Spec{
		OracleID:     "cd3d1e5d-db12-4878-92a2-23eb8d23eaf3",
		Name:         "Blazing Torch",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: blazingTorchGrant,
			Activated: []ActivatedAbility{{
				Label:   "{T}, Sacrifice Blazing Torch: Blazing Torch deals 2 damage to any target.",
				Purpose: game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove},
				Cost:    Plus(TapCost(), SacrificeTheGrantor()),
				Targets: TargetAny(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if len(item.Targets) == 0 {
						return nil
					}
					return DealDamage{Source: ctx.Grantor(), Target: item.Targets[0].ID, Amount: 2}.Apply(ctx)
				},
			}},
			Text: "{T}, Sacrifice Blazing Torch: Blazing Torch deals 2 damage to any target.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(blazingTorchGrant)},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnAttached(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.HasSubtypeWord("Vampire") || c.HasSubtypeWord("Zombie")
			}, "Vampires and Zombies"),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}
