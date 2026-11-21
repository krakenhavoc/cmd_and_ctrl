package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tannuk, Memorial Ensign — Legendary Creature — Kavu Pilot {1}{R}{G}, 2/4
// (Edge of Eternities, #4461):
//
//	"Landfall — Whenever a land you control enters, Tannuk deals 1 damage
//	 to each opponent. If this is the second time this ability has
//	 resolved this turn, draw a card."
//
// Captain Michael Burnham's landfall with a draw in place of her counter.
// The resolution tally (Game.ResolvedThisTurn, keyed on this trigger's
// stack label) counts the resolution in progress, so the second
// resolution reads exactly 2, and the first and third draw nothing. Per
// object (CR 400.7): a Tannuk that left and came back this turn starts
// again. The damage is dealt even if Tannuk has left the battlefield, as
// it last existed (CR 608.2h), and the draw needs nothing of it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76cb1912-3791-44b8-b594-b92dab21bafc",
		Name:         "Tannuk, Memorial Ensign",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying(tannukMemorialEnsignLabel, game.LandfallPayoff{DamageEachOpponent: 1}, tannukMemorialEnsignLandfall),
		},
	})
}

const tannukMemorialEnsignLabel = "Tannuk, Memorial Ensign — landfall: 1 damage to each opponent, and a card the second time this turn"

func tannukMemorialEnsignLandfall(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.DamageEachOpponent{N: 1}).Apply(ctx); err != nil {
		return err
	}
	if g.ResolvedThisTurn(item.SourceCardID, tannukMemorialEnsignLabel) != 2 {
		return nil
	}
	return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
}
