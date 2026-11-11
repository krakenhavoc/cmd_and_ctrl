package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Topiary Stomper — Creature — Plant Dinosaur {1}{G}{G}, 4/4 (Lost Caverns of
// Ixalan Commander, #5368):
//
//	"Vigilance
//	 When this creature enters, search your library for a basic land card, put it
//	 onto the battlefield tapped, then shuffle.
//	 This creature can't attack or block unless you control seven or more
//	 lands."
//
// The search is Rampant Growth's. The restriction is Hazoret's shape (a layer-6
// self static that sets CantAttack and CantBlock while a condition fails), here
// "you control fewer than seven lands", counted live every time the layers are
// recomputed, so the Stomper can attack the turn the seventh land arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e3fc10d2-2733-4265-8275-112c285fd11a",
		Name:            "Topiary Stomper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Purpose:         game.Purpose{Lands: 1},
		Static: []game.StaticAbility{{
			Label: "Can't attack or block unless you control seven or more lands.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfOnly(target, g, source) && b10LandsControlled(g, source.Controller) < 7
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Topiary Stomper — search your library for a basic land card, put it onto the battlefield tapped, then shuffle",
				cmdr2B15SearchBasicLand("Topiary Stomper — a basic land card, onto the battlefield tapped", true)),
		},
	})
}
