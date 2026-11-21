package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reap the Seagraf — Sorcery {2}{B} (Dark Ascension, #5571):
//
//	"Create a 2/2 black Zombie creature token.
//	 Flashback {4}{U} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Moan of the Unhallowed with one token and a blue flashback cost (CR
// 702.34a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "95ff3ba5-cbd2-4723-9498-3fc6df5be973",
		Name:             "Reap the Seagraf",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1, CreatureTokens: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.CostWithPurpose(effects.Flashback("{4}{U}"), game.Purpose{Tokens: 1, CreatureTokens: 1})},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("2/2 black Zombie"), N: 1}.Apply(ctx)
		},
	})
}
