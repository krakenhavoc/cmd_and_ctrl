package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdrB5Dragon is the Dragon creature type, read through the effective
// subtype list so a changeling or a granted type counts.
func cmdrB5Dragon() CardPredicate { return OfCreatureType("Dragon") }

// cmdrB5CountBattlefield counts the battlefield permanents, of any
// controller, that pass pred, read through the layered view.
func cmdrB5CountBattlefield(g *game.Game, pred func(game.Card) bool) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if pred(c) {
			n++
		}
	}
	return n
}

// cmdrB5CountYouControl counts the permanents `controller` controls that
// pass pred.
func cmdrB5CountYouControl(g *game.Game, controller uuid.UUID, pred func(game.Card) bool) int {
	return cmdrB5CountBattlefield(g, func(c game.Card) bool {
		return c.Controller == controller && pred(c)
	})
}

// cmdrB5CountAllGraveyards counts the cards in every graveyard that pass
// pred.
func cmdrB5CountAllGraveyards(g *game.Game, pred func(game.Card) bool) int {
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if pred(c) {
				n++
			}
		}
	}
	return n
}

// cmdrB5DivideGreatestPower is "X damage divided as you choose, where X is
// the greatest power among creatures you control as you cast this spell"
// (Monstrous Onslaught): the amount is read as the targets are announced,
// so it is fixed for the cast (CR 601.2d, 601.2i).
var cmdrB5DivideGreatestPower = game.RegisterDivideAmount("greatest-power-you-control",
	func(g *game.Game, a game.DivideAmountArgs) game.DivideSpec {
		return game.DivideSpec{Total: b42GreatestPowerControlledBy(g, a.Controller)}
	})
