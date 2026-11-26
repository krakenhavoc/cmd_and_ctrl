package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jace's Sentinel — Creature — Merfolk Warrior {1}{U}, 1/3 (Ixalan):
//
//	"As long as you control a Jace planeswalker, this creature gets +1/+0 and
//	 can't be blocked."
//
// Rowan's Battleguard's layer 7c static, and Yuan-Ti Malison's block rule with
// the Jace test as the condition: both are read live, the +1/+0 on every
// recompute and the block rule as each block is checked. "Jace" is a
// planeswalker type, matched on the permanent's subtypes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfd2ef5d-618d-4062-9cb2-9dbef5432865",
		Name:         "Jace's Sentinel",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "As long as you control a Jace planeswalker, this creature gets +1/+0.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.SelfCreature(target, g, source) && effects.OldB222YouControlAPlaneswalker(g, source.Controller, "Jace")
			},
			Apply: effects.PumpBy(1, 0),
		}},
		BlockRules: []game.BlockRule{effects.CantBeBlockedWhile(effects.OnSelf(), func(g *game.Game, you uuid.UUID, _ game.Card) bool {
			return effects.OldB222YouControlAPlaneswalker(g, you, "Jace")
		})},
	})
}
