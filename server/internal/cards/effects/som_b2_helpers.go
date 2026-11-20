package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// som_b2_helpers.go — the shapes Scars of Mirrodin, batch 2 (#5577)
// prints more than once. Append-only.

// somB2MetalcraftKeywords is "Metalcraft — This creature has <keyword>
// as long as you control three or more artifacts." (Auriok Edgewright,
// Snapsail Glider): a layer 6 self grant on somB1Metalcraft's gate. Like
// the pump, it is not locked in (CR 611.3a): it comes and goes as the
// count crosses three.
func somB2MetalcraftKeywords(label string, keywords ...string) game.StaticAbility {
	s := b16GrantKeywords(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfOnly(target, g, source) && somB1Metalcraft(g, source)
	}, keywords...)
	s.Label = label
	return s
}

// somB2SpellbombDraw is the Spellbombs' second ability, "When this
// artifact is put into a graveyard from the battlefield, you may pay
// <cost>. If you do, draw a card." (Nihil Spellbomb's shape, with its
// paid branch). It fires however the Spellbomb got there, its own
// sacrifice included; declining, or a pool that can't pay, draws nothing.
func somB2SpellbombDraw(name, cost string) game.TriggeredAbility {
	return WhenThisDies(name+" — you may pay "+cost+"; if you do, draw a card",
		func(g *game.Game, item *game.StackItem) error {
			return MayPay{
				Chooser:  item.Controller,
				Cost:     cost,
				Question: "Pay " + cost + " to draw a card?",
				Then:     rprB15SpellbombPaid,
			}.Apply(NewContext(g, item))
		})
}
