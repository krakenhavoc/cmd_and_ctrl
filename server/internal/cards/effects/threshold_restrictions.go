package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// threshold_restrictions.go — the "can't" half of a threshold creature
// (Odyssey, #5613). The P/T and keyword halves are fraThresholdSelfPT and
// fraThresholdSelfKeywords.
//
// Append-only: add a constructor, never change what one means.

// thresholdSelfCantBlock is "Threshold — As long as there are seven or
// more cards in your graveyard, this creature … can't block" (Frightcrawler,
// Childhood Horror). It is Patchwork Beastie's shape: a layer 6 static on
// the creature itself that sets CantBlock while the threshold holds, read
// live as blockers are declared (CR 509.1b), so the restriction comes and
// goes with the seventh card. The creature can still attack.
func thresholdSelfCantBlock() game.StaticAbility {
	return game.StaticAbility{
		Label: "Can't block while there are seven or more cards in your graveyard.",
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return selfOnly(target, g, source) && fraThresholdMet(g, source.Controller)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Restrictions |= game.CantBlock
		},
	}
}
