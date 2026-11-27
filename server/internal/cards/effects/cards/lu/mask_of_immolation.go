package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mask of Immolation — Artifact — Equipment {1}{R} (Core Set 2020):
//
//	"When this Equipment enters, create a 1/1 red Elemental creature token,
//	 then attach this Equipment to it.
//	 Equipped creature has "Sacrifice this creature: It deals 1 damage to any
//	 target."
//	 Equip {2} ({2}: Attach to target creature you control. Equip only as a
//	 sorcery.)"
//
// Ancestral Blade's enters trigger over a 1/1 red Elemental. The grant is
// Blazing Torch's ability bundle on the attached creature: the creature is
// sacrificed to pay (CR 118.12), and the damage is dealt by it as it last
// existed (CR 608.2h), so a lifelinking or deathtouching creature keeps those
// on the damage. It works for any creature the Mask is attached to, the
// Elemental or not, and stops when the Mask is unattached.
//
// No simplification.
const maskOfImmolationGrant = "mask-of-immolation/deal-1"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd86d11f-0bd8-49a9-936d-e407ea2c1818",
		Name:         "Mask of Immolation",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: maskOfImmolationGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "Sacrifice this creature: It deals 1 damage to any target.",
				Purpose: game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 1)), Answers: game.AnswerRemove},
				Cost:    effects.SacrificeThis(),
				Targets: effects.TargetAny(),
				Effect:  effects.OldB466SourceDealsOneDamage,
			}},
			Text: "Sacrifice this creature: It deals 1 damage to any target.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(maskOfImmolationGrant)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mask of Immolation — create a 1/1 red Elemental creature token, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokenAndAttachSource(g, item, "1/1 red Elemental", true)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
