package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harbor Serpent — Creature — Serpent {4}{U}{U}, 5/5 (Magic 2011, #5579):
//
//	"Islandwalk
//	 This creature can't attack unless there are five or more Islands on
//	 the battlefield."
//
// Islandwalk is the printed keyword the block check reads. The attack
// restriction is Chief Warg's Company's shape: a layer 6 static on the
// Serpent that sets CantAttack while fewer than five Islands are on the
// battlefield. Every player's Islands count, not only its controller's,
// and the count is read live on each pass, so a land that became an
// Island counts. "Island" is a word of the card's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "599fb8b1-8750-4832-869a-cddc3784087a",
		Name:            "Harbor Serpent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		Static: []game.StaticAbility{{
			Label: "Can't attack unless there are five or more Islands on the battlefield.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.SelfOnly(target, g, source) && m11IslandsOnTheBattlefield(g, source) < 5
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack
			},
		}},
	})
}

// m11IslandsOnTheBattlefield is the number of Islands on the battlefield,
// whoever controls them. It walks g.Battlefield.Cards directly, as the
// layer pass that reads it may hold only the read lock (permanentsYouControl).
func m11IslandsOnTheBattlefield(g *game.Game, source *game.Card) int {
	if g == nil || g.Battlefield == nil {
		return 0
	}
	island := source.TextSubtype("Island")
	n := 0
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].HasSubtype(island) {
			n++
		}
	}
	return n
}
