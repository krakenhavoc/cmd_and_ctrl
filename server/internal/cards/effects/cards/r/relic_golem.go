package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Relic Golem — Artifact Creature — Golem {3}, 6/6 (Zendikar Rising, #5508):
//
//	"This creature can't attack or block unless an opponent has eight or
//	 more cards in their graveyard.
//	 {2}, {T}: Target player mills two cards."
//
// The restriction is a layer 6 static, on while no opponent has eight cards
// in their graveyard (Expedition Lookout's live read, so it flips as the
// graveyards fill). The mill is Scavenger's Talent's: a target that is no
// longer legal as it resolves mills nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14d9bb24-df7b-4f41-aeca-aad639b280db",
		Name:         "Relic Golem",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "This creature can't attack or block unless an opponent has eight or more cards in their graveyard.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.SelfOnly(target, g, source) && !effects.S82OpponentHasGraveyardCards(g, source.Controller, 8)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Target player mills two cards.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.TargetPlayerMillsTwo,
		}},
	})
}
