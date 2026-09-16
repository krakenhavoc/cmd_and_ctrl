package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vampiric Rites — Enchantment {B}:
//
//	"{1}{B}, Sacrifice a creature: You gain 1 life and draw a card."
//
// A one-mana enchantment that turns every creature into a card. Like
// Bastion of Remembrance, the outlet itself survives creature
// removal, so it keeps working through the wipe that feeds it.
//
// The cost pairs mana with SACRIFICE-ANOTHER (SacrificeACreature),
// distinct from Mind Stone's sacrifice-self: the controller picks
// which creature from a list at announce, and the engine enforces
// CR 701.21a (it must be one they control).
func init() {
	Register(Spec{
		OracleID: "660de988-b6fb-4f36-8006-42af3e7f908d",
		Name:     "Vampiric Rites",
		Activated: []ActivatedAbility{{
			Label: "{1}{B}, Sacrifice a creature: You gain 1 life and draw a card.",
			Cost: func() game.AbilityCost {
				c := SacrificeACreature()
				c.Mana = "{1}{B}"
				return c
			}(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (GainLife{Player: item.Controller, Amount: 1}).Apply(ctx); err != nil {
					return err
				}
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			},
		}},
	})
}
