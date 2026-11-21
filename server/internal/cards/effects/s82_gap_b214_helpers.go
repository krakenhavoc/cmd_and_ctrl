package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b214_helpers.go — shared shapes for GAP batch 214 (#5948).

// GapB214LifeAtMostHalfStarting is "your life total is less than or equal to
// half your starting life total" (the Baldur's Gate gods): the source's
// controller's life against the table's starting life, compared as life*2 <=
// starting so an odd total has no rounding.
func GapB214LifeAtMostHalfStarting(g *game.Game, source *game.Card) bool {
	if g == nil || source == nil {
		return false
	}
	for _, p := range g.Seats {
		if p != nil && p.ID == source.Controller {
			return p.Life*2 <= StartingLifeOf(g)
		}
	}
	return false
}

// GapB214SelfIndestructibleAtHalfLife is "As long as your life total is less
// than or equal to half your starting life total, <this> has indestructible"
// (Bhaal, Myrkul, Lord of Bones and batch 215's Bane, Lord of Darkness): a
// layer 6 keyword grant (CR 613.1f) that re-reads on every life change.
func GapB214SelfIndestructibleAtHalfLife(name string) game.StaticAbility {
	ab := LifeGatedKeyword(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && GapB214LifeAtMostHalfStarting(g, source)
	}, "indestructible")
	ab.Label = "As long as your life total is less than or equal to half your starting life total, " + name + " has indestructible."
	return ab
}
