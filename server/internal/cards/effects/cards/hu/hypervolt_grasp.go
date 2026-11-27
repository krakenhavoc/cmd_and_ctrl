package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hypervolt Grasp — Enchantment — Aura {2}{R} (Guildpact):
//
//	"Enchant creature
//	 Enchanted creature has "{T}: This creature deals 1 damage to any
//	 target."
//	 {1}{U}: Return this Aura to its owner's hand."
//
// Burning Anger's granted tap ability with 1 damage (ADR 0093): the
// enchanted creature's controller activates it and the creature is the
// source. The bounce is Flickering Ward's: a mana-cost activation of the Aura
// itself, with no target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "269a2b53-52d7-4032-ba62-7b7742e7a159",
		Name:         "Hypervolt Grasp",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: hypervoltGraspGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This creature deals 1 damage to any target.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Effect:  effects.DealDamageToTheTarget(1),
			}},
			Text: "{T}: This creature deals 1 damage to any target.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(hypervoltGraspGrant)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: Return this Aura to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BounceToHand{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// hypervoltGraspGrant names the granted ability's bundle: an on-disk
// identity.
const hypervoltGraspGrant = "hypervolt-grasp/tap-ping-any-target"
