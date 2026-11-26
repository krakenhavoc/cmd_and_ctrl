package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prickleboar — Creature — Boar {4}{R}, 3/3 (Magic Origins):
//
//	"During your turn, this creature gets +2/+0 and has first strike."
//
// Dapper Shieldmate's turn-gated pump and Bearer of Glory's turn-gated
// first strike, one static each: both read the active player at each
// layer recompute, so the creature is a 3/3 on every other turn and a
// 5/3 first striker on its controller's.
//
// No simplification.
func init() {
	mineAndMyTurn := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.ThisPermanent(target, g, source) && effects.IsActivePlayer(g, source.Controller)
	}
	effects.Register(effects.Spec{
		OracleID:     "a023deaa-c393-44f9-9c9c-6bb9aae05c1a",
		Name:         "Prickleboar",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "During your turn, this creature gets +2/+0.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: mineAndMyTurn,
				Apply:     effects.PumpBy(2, 0),
			},
			effects.KeywordGrant(mineAndMyTurn, "first strike"),
		},
	})
}
