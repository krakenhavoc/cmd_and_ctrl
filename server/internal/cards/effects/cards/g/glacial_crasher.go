package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glacial Crasher — Creature — Elemental {4}{U}{U}, 5/5 (Magic 2015, #5556):
//
//	"Trample
//	 This creature can't attack unless there is a Mountain on the
//	 battlefield."
//
// Harbor Serpent's shape with a count of one: a layer 6 static on the
// Crasher that sets CantAttack while no Mountain is on the battlefield.
// Every player's Mountains count, and the count is read live on each pass,
// so a land that became a Mountain counts. "Mountain" is a word of the
// card's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "42441362-d638-423a-8928-ab3fd9b2d4f2",
		Name:            "Glacial Crasher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label: "Can't attack unless there is a Mountain on the battlefield.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.SelfOnly(target, g, source) && !glacialCrasherMountainInPlay(g, source)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack
			},
		}},
	})
}

// glacialCrasherMountainInPlay reports whether any Mountain is on the
// battlefield, whoever controls it. It walks g.Battlefield.Cards directly, as
// the layer pass that reads it may hold only the read lock.
func glacialCrasherMountainInPlay(g *game.Game, source *game.Card) bool {
	if g == nil || g.Battlefield == nil {
		return false
	}
	mountain := source.TextSubtype("Mountain")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].HasSubtype(mountain) {
			return true
		}
	}
	return false
}
