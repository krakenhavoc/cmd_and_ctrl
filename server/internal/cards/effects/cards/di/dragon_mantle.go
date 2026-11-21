package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const dragonMantleGrant = "dragon-mantle/firebreathing"

// Dragon Mantle — Enchantment — Aura {R} (Enchanting Tales):
//
//	"Enchant creature
//	 When this Aura enters, draw a card.
//	 Enchanted creature has "{R}: This creature gets +1/+0 until end of
//	 turn.""
//
// Burning Anger's grant shape (ADR 0093): the granted row is the enchanted
// creature's own ability, so "this creature" is the host, and Shivan Dragon's
// firebreathing body is the effect. The cantrip is an ordinary ETB trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0cc5aaa1-07bf-4872-ac72-303cef9fb143",
		Name:         "Dragon Mantle",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Grants: []effects.AbilityGrant{{
			Key: dragonMantleGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{R}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Dragon Mantle — +1/+0"}.Apply(ctx)
				},
			}},
			Text: "{R}: This creature gets +1/+0 until end of turn.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(dragonMantleGrant)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dragon Mantle — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
