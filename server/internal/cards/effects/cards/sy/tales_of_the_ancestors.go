package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tales of the Ancestors — Sorcery {3}{U} (Kaldheim Commander, #5506):
//
//	"Each player with fewer cards in hand than the player with the most
//	 cards in hand draws cards equal to the difference.
//	 Foretell {1}{U}"
//
// The most is read once as the spell resolves, before anyone draws (CR
// 608.2e), and each player short of it draws the gap, in APNAP order. A
// player tied for the most draws nothing. Foretell is the engine's special
// action.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "49e3c192-862f-483e-abd1-e22863ea88dd",
		Name:           "Tales of the Ancestors",
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Foretell("{1}{U}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			g := ctx.Game
			size := func(id uuid.UUID) int {
				p := g.PlayerByIDForEffect(id)
				if p == nil || p.Hand == nil {
					return 0
				}
				return p.Hand.Size()
			}
			players := effects.ApnapPlayers(g)
			most := 0
			for _, id := range players {
				if n := size(id); n > most {
					most = n
				}
			}
			gaps := make(map[uuid.UUID]int, len(players))
			for _, id := range players {
				gaps[id] = most - size(id)
			}
			for _, id := range players {
				if gaps[id] > 0 {
					if err := (effects.DrawCards{Player: id, N: gaps[id]}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
