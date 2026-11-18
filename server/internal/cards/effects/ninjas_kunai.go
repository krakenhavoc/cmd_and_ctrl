package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ninja's Kunai — Artifact — Equipment {1} (Kamigawa: Neon Dynasty,
// #5493):
//
//	"Equipped creature has "{1}, {T}, Sacrifice Ninja's Kunai: Ninja's
//	 Kunai deals 3 damage to any target."
//	 Equip {1}"
//
// An ADR 0093 grant (Deconstruction Hammer's shape): the ability is the
// equipped creature's own, so the creature taps to pay for it and CR 302.6
// applies, and the Kunai is sacrificed as part of the cost
// (SacrificeTheGrantor). The damage is dealt by the Kunai, as it last
// existed, so it is the Kunai's colourless artifact damage and not the
// creature's.
//
// No simplification.
const ninjasKunaiGrant = "ninjas-kunai/deal-3"

func init() {
	Register(Spec{
		OracleID:     "fa5aefaf-432f-41d0-8e54-0e8b7bc7ba98",
		Name:         "Ninja's Kunai",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: ninjasKunaiGrant,
			Activated: []ActivatedAbility{{
				Label:   "{1}, {T}, Sacrifice Ninja's Kunai: Ninja's Kunai deals 3 damage to any target.",
				Purpose: game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove},
				Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeTheGrantor()),
				Targets: TargetAny(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if len(item.Targets) == 0 {
						return nil
					}
					return DealDamage{Source: ctx.Grantor(), Target: item.Targets[0].ID, Amount: 3}.Apply(ctx)
				},
			}},
			Text: "{1}, {T}, Sacrifice Ninja's Kunai: Ninja's Kunai deals 3 damage to any target.",
		}},
		Static:    []game.StaticAbility{GrantAbilitiesToAttached(ninjasKunaiGrant)},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}
