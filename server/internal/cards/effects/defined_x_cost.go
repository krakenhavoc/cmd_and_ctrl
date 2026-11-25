package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// defined_x_cost.go — an activated ability's {X} that its text defines
// (CR 107.3c, #6240):
//
//	"{X}, {T}: Create a token that's a copy of the exiled card. X is the
//	 mana value of that card."            (Soul Foundry, Prototype Portal)
//
// The activator doesn't choose X: the cost is generic mana equal to the
// mana value of the card exiled with this permanent, priced by the engine
// (game.AbilityCost.ManaValueOfExiledWith), and nothing asks for an X.
//
//	Cost: Plus(XIsTheManaValueOfTheExiledCard(), TapCost()),

// XIsTheManaValueOfTheExiledCard is that mana component.
func XIsTheManaValueOfTheExiledCard() game.AbilityCost {
	return game.AbilityCost{ManaValueOfExiledWith: true}
}
