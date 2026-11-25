package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tezzeret's Strider — Artifact Creature — Golem {3}, 3/1 (Core Set 2019, S82 OLD batch 200, #5527):
//
//	"As long as you control a Tezzeret planeswalker, this creature has
//	 menace."
//
// Karakyk Guardian's keyword grant: a layer 6 keyword grant to itself while you control
// a planeswalker with the Tezzeret subtype (an opponent's does not count).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cac79a2b-1de3-4461-95ac-fa0f1129cd6b",
		Name:         "Tezzeret's Strider",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Has menace as long as you control a Tezzeret planeswalker.",
			Layer:     game.Layer6Ability,
			AppliesTo: effects.SelfWhileYouControlA(effects.And(effects.Planeswalker(), effects.Subtype("Tezzeret"))),
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				effects.AppendKeywordsTo(c, []string{"menace"})
			},
		}},
	})
}
