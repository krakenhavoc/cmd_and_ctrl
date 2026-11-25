package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nature's Resurgence — Sorcery {2}{G}{G} (Weatherlight, #5629):
//
//	"Each player draws a card for each creature card in their graveyard."
//
// Every graveyard is counted first, once (CR 608.2h), then the players
// draw in turn order from the active player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "276f49ee-7cf3-4a4d-9b14-74aedcbef69f",
		Name:         "Nature's Resurgence",
		Completeness: effects.CompletenessFull,
		OnResolve:    naturesResurgenceDraws,
	})
}

// naturesResurgenceDraws counts each player's creature cards in their
// graveyard before anyone draws, then has each player still in the game
// draw that many, from the active player in turn order.
func naturesResurgenceDraws(_ *game.StackItem, ctx *effects.Context) error {
	g := ctx.Game
	n := len(g.Seats)
	owed := make([]int, n)
	for i := 0; i < n; i++ {
		if p := g.Seats[g.SeatInTurnOrderForEffect(g.Turn.ActiveSeat, i)]; p != nil && !p.Eliminated {
			owed[i] = effects.B11CreatureCardsInGraveyard(g, p.ID)
		}
	}
	for i := 0; i < n; i++ {
		if p := g.Seats[g.SeatInTurnOrderForEffect(g.Turn.ActiveSeat, i)]; p != nil && !p.Eliminated {
			if err := (effects.DrawCards{Player: p.ID, N: owed[i]}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}
