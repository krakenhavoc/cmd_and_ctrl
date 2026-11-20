package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// arb_b1_helpers.go — the shapes Alara Reborn, batch 1 (#5586) prints
// more than once. Append-only.

// arbB1AnotherMulticolored is "you control another multicolored
// permanent" (the Alara Reborn blades): some permanent other than the
// source, controlled by the source's controller, with two or more
// colours after the layers. It walks g.Battlefield.Cards directly
// because it runs inside a layer recompute (see sharedCreatureTypeCount),
// and the colours it reads are the ones layer 5 has already written for
// the pass.
func arbB1AnotherMulticolored(g *game.Game, source *game.Card) bool {
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID == source.InstanceID || c.Controller != source.Controller {
			continue
		}
		if len(c.EffectiveColors()) >= 2 {
			return true
		}
	}
	return false
}

// arbB1SelfWhileAnotherMulticolored is the AppliesTo of "As long as you
// control another multicolored permanent, this creature …": the source
// alone, while the condition holds.
func arbB1SelfWhileAnotherMulticolored(target *game.Card, g *game.Game, source *game.Card) bool {
	return selfCreature(target, g, source) && arbB1AnotherMulticolored(g, source)
}

// arbB1Blade is the cycle's whole static: "As long as you control another
// multicolored permanent, this creature gets +1/+1 and has <keyword>." A
// layer 7c +1/+1 and a layer 6 grant (CR 613.1f), both read live, so the
// bonus ends the moment the last other multicolored permanent leaves or
// loses a colour.
func arbB1Blade(keyword string) []game.StaticAbility {
	grant := KeywordGrant(arbB1SelfWhileAnotherMulticolored, keyword)
	grant.Label = "As long as you control another multicolored permanent, this creature has " + keyword + "."
	return []game.StaticAbility{
		{
			Label:     "As long as you control another multicolored permanent, this creature gets +1/+1.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: arbB1SelfWhileAnotherMulticolored,
			Apply:     pumpBy(1, 1),
		},
		grant,
	}
}
