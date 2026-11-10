package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Final Days — Sorcery {2}{B}{B} (Final Fantasy, #4462):
//
//	"Create two tapped 2/2 black Horror creature tokens. If this spell
//	 was cast from a graveyard, instead create X of those tokens, where
//	 X is the number of creature cards in your graveyard.
//	 Flashback {4}{B}{B}"
//
// X is counted once, as the spell resolves (the ruling), from the
// controller's graveyard; the spell itself is on the stack, not in it.
// Cast from the graveyard (the announce record's zone) makes X tokens
// and not two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f2b88031-bfb7-46c4-abdd-3b6b5f5acfa4",
		Name:             "The Final Days",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 2, CreatureTokens: 2},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{4}{B}{B}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := 2
			if item.CastFromZone == game.ZoneGraveyard {
				n = b11CreatureCardsInGraveyard(ctx.Game, ctx.Controller())
			}
			if n <= 0 {
				return nil
			}
			return CreateTokenAdvanced{Spec: Token(TokenCard("2/2 black Horror")).EntersTapped(), N: n}.Apply(ctx)
		},
	})
}
