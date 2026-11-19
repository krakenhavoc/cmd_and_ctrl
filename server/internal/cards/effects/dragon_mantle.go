package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0cc5aaa1-07bf-4872-ac72-303cef9fb143",
		Name:         "Dragon Mantle",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Grants: []AbilityGrant{{
			Key: dragonMantleGrant,
			Activated: []ActivatedAbility{{
				Label:   "{R}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Dragon Mantle — +1/+0"}.Apply(ctx)
				},
			}},
			Text: "{R}: This creature gets +1/+0 until end of turn.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(dragonMantleGrant)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dragon Mantle — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
