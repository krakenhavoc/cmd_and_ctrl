package cardsg

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gor Muldrak, Amphinologist — Legendary Creature — Human Scout {1}{G}{U},
// 3/2 (Commander Legends):
//
//	"You and permanents you control have protection from Salamanders.
//	 At the beginning of your end step, each player who controls the fewest
//	 creatures creates a 4/3 blue Salamander Warrior creature token."
//
// Serra's Emissary's shape: a layer-6 static gives each permanent you control
// the protection (a Salamander Warrior token you control is a Salamander, so it
// can't damage, enchant, block or be blocked by your permanents), and the player
// half is a player keyword. The end-step ability counts creatures as it
// resolves; every player tied for the fewest makes one token, an eliminated
// player is not counted.
//
// No simplification.
func init() {
	const prot = "protection from Salamanders"
	effects.Register(effects.Spec{
		OracleID:       "9abdd53a-aa62-447b-9525-02a2227cc44b",
		Name:           "Gor Muldrak, Amphinologist",
		Completeness:   effects.CompletenessFull,
		PlayerKeywords: []string{prot},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "You and permanents you control have protection from Salamanders.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !slices.Contains(c.Abilities, prot) {
					c.Abilities = append(c.Abilities, prot)
				}
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.ByYou,
				"Gor Muldrak — each player who controls the fewest creatures creates a 4/3 blue Salamander Warrior creature token",
				effects.Do(effects.S82OldB083FewestCreaturesEachMakes{Template: effects.TokenCard("4/3 blue Salamander Warrior")})),
		},
	})
}
