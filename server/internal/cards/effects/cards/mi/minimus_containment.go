package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const minimusContainmentTreasure = "minimus-containment/treasure"

// Minimus Containment — Enchantment — Aura {2}{W} (Adventures in the Forgotten
// Realms):
//
//	"Enchant nonland permanent
//	 Enchanted permanent is a Treasure artifact with '{T}, Sacrifice this
//	 artifact: Add one mana of any color,' and it loses all other abilities.
//	 (If it was a creature, it's no longer a creature.)"
//
// Sugar Coat's shape: one continuous effect from layer 4 that sets the card
// types to Artifact and the subtypes to Treasure (CR 205.1a), then layer 6
// removes every ability and grants the Treasure's own. The host keeps its
// colours, name and counters. Mana from it is Treasure mana (the source kind
// is read from the effective subtype).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a5e6223-ca31-4605-9c33-9ce492aaf556",
		Name:         "Minimus Containment",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant nonland permanent", effects.Nonland()),
		Grants: []effects.AbilityGrant{{
			Key: minimusContainmentTreasure,
			Mana: []effects.ManaAbility{{
				Cost:     effects.ManaAbilityCost{Tap: true, Sacrifice: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}, Sacrifice this artifact: Add one mana of any color",
			}},
			Text: "{T}, Sacrifice this artifact: Add one mana of any color.",
		}},
		Static: []game.StaticAbility{
			{
				Label:                 "Enchanted permanent is a Treasure artifact and loses all other card types.",
				Layer:                 game.Layer4Type,
				ContinuesAfterRemoval: true,
				AppliesTo:             effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					c.Types = []string{"Artifact"}
					c.SetSubtypes(nil)
					c.Subtypes = append(c.Subtypes, source.TextSubtypes("Treasure")...)
				},
			},
			effects.LosesAllAbilitiesAndHas(effects.AttachedToSource, minimusContainmentTreasure),
		},
	})
}
