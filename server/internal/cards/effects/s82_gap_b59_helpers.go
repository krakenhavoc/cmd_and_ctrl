package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// costsLessPerCardDrawnThisTurn is "This spell costs {1} less to cast for
// each card you've drawn this turn" (Deem Inferior), read at CR 601.2f off the
// caster's per-turn tally of cards drawn.
func CostsLessPerCardDrawnThisTurn() game.CostModifier {
	return CostsLessEach(cardsDrawnThisTurnByCaster, "This spell costs {1} less to cast for each card you've drawn this turn.")
}

func cardsDrawnThisTurnByCaster(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	return q.Game.TurnTallyFor(q.Controller).CardsDrawn
}
