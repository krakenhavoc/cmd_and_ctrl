package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghoulcaller's Harvest — Sorcery {B}{G} (Midnight Hunt):
//
//	"Create X 2/2 black Zombie creature tokens with decayed, where X is
//	 half the number of creature cards in your graveyard, rounded up.
//	 Flashback {3}{B}{G}"
//
// X is counted as the spell resolves, from the caster's own graveyard
// (the card itself is a sorcery, and a flashed-back one is on the stack,
// so it never counts). Rounded up: (n+1)/2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "bbef4d8e-5861-4d50-9589-14b46c71e9af",
		Name:             "Ghoulcaller's Harvest",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1, CreatureTokens: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{B}{G}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := (b11CreatureCardsInGraveyard(ctx.Game, ctx.Controller()) + 1) / 2
			return CreateToken{Template: TokenCard("2/2 black Zombie with decayed"), N: n}.Apply(ctx)
		},
	})
}
