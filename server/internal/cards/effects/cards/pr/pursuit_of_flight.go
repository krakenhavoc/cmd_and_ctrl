package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pursuit of Flight — Enchantment — Aura {1}{R} (Return to Ravnica, S82 OLD):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has "{U}: This creature gains flying
//	 until end of turn.""
//
// Deviant Glee's pump-and-grant with +2/+2 and flying for {U}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cdce8b86-81e3-443c-af91-f6b4f62f5b59",
		Name:         "Pursuit of Flight",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: pursuitOfFlightGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{U}: This creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{U}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"}, Label: "Pursuit of Flight — flying until end of turn"}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{U}: This creature gains flying until end of turn.",
		}},
		Static: []game.StaticAbility{effects.PumpAttached(2, 2), effects.GrantAbilitiesToAttached(pursuitOfFlightGrant)},
	})
}

// pursuitOfFlightGrant is the granted ability's bundle key: an on-disk
// identity, never renamed, never reused.
const pursuitOfFlightGrant = "pursuit-of-flight/gain-flying"
