package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strike It Rich — Sorcery {R}:
//
//	"Create a Treasure token. (It's an artifact with '{T}, Sacrifice
//	 this token: Add one mana of any color.')
//	 Flashback {2}{R} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// One Treasure now, and a second one later out of the graveyard for a
// little more mana — the whole card. The token is the real Treasure
// with its real sacrifice-for-mana ability, not a stand-in.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "c34c17c7-3827-49a2-be25-67f44fdfe150",
		Name:             "Strike It Rich",
		Purpose:          game.Purpose{Tokens: 1},
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{2}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TreasureToken(), N: 1}.Apply(ctx)
		},
	})
}
