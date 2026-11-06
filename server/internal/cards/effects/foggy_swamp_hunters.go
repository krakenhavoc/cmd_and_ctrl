package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foggy Swamp Hunters — Creature — Human Ranger Ally {3}{B}, 3/4:
//
//	"As long as you've drawn two or more cards this turn, this creature
//	 has lifelink and menace."
//
// A layer 6 grant to itself, read off the turn's draw tally for its
// controller (Lake-town Toymaker's count). The tally is not on the
// battlefield, so the static declares DependsOnHandSize: a draw is a
// card entering a hand, which drops the cached layers at once, and the
// turn's advance clears both the tally and the cache.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "1941cf7a-ad53-4bce-beea-104951619a30",
		Name:         "Foggy Swamp Hunters",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			foggySwampHuntersGrant("lifelink"),
			foggySwampHuntersGrant("menace"),
		},
	})
}

// foggySwampHuntersGrant is "this creature has <keyword> as long as
// you've drawn two or more cards this turn".
func foggySwampHuntersGrant(keyword string) game.StaticAbility {
	ab := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && g.TurnTallyFor(source.Controller).CardsDrawn >= 2
	}, keyword)
	ab.Label = "Has " + keyword + " as long as you've drawn two or more cards this turn."
	ab.DependsOnHandSize = true
	return ab
}
