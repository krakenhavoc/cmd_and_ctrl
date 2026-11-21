package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// createThreeSoldiers is "create three 1/1 white Soldier creature tokens" as
// a planeswalker ability's effect (Elspeth, Sun's Champion's +1, Elspeth
// Tirel's −2).
func CreateThreeSoldiers(g *game.Game, item *game.StackItem) error {
	return CreateToken{
		Controller: item.Controller,
		Template:   TokenCard("1/1 white Soldier"),
		N:          3,
	}.Apply(NewContext(g, item))
}
