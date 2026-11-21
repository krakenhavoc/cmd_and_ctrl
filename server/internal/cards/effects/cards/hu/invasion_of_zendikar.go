package cardshu

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invasion of Zendikar // Awakened Skyclave — March of the Machine (#5477).
//
// Front, Battle — Siege 3, {3}{G}: "When this Siege enters, search your
// library for up to two basic land cards, put them onto the battlefield
// tapped, then shuffle." Back, Awakened Skyclave, Creature — Elemental 4/4:
// "Vigilance, haste. As long as this creature is on the battlefield, it's a
// land in addition to its other types. {T}: Add one mana of any color."
//
// The land type is a layer 4 effect on the battlefield only, so the card is
// a plain creature when it is cast transformed (it is on the stack), and a
// land creature once it resolves.
//
// No simplification.
const InvasionOfZendikarOracleID = "4b1874af-4ea5-4e22-a4d4-e718d75fe95e"

func init() {
	effects.Register(effects.Spec{
		OracleID:     InvasionOfZendikarOracleID,
		Name:         "Invasion of Zendikar",
		Completeness: effects.CompletenessFull,
		Battle:       &effects.BattleSpec{Defense: 3, Subtype: effects.BattleSubtypeSiege},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Invasion of Zendikar — search for up to two basic lands, put them onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     effects.IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         2,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        "Invasion of Zendikar — up to two basic lands",
					}.Apply(effects.NewContext(g, item))
				}),
			effects.DefeatedTrigger("Invasion of Zendikar — defeated: exile it, then cast Awakened Skyclave", effects.SiegeDefeated()),
		},
		Purpose: game.Purpose{Lands: 2},
	})
	effects.Register(effects.Spec{
		OracleID:        InvasionOfZendikarOracleID + "#1",
		Name:            "Awakened Skyclave",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "haste"},
		Static: []game.StaticAbility{{
			Layer:     game.Layer4Type,
			Label:     "As long as this creature is on the battlefield, it's a land in addition to its other types.",
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !slices.Contains(c.Types, "Land") {
					c.Types = append(c.Types, "Land")
				}
			},
		}},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
