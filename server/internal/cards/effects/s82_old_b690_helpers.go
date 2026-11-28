package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b690_helpers.go — shared pieces of the Mercadian Masques damage cards
// of S82 OLD batch 690 (#5619).

// OldB690DivideCreaturesOnBattlefield is Volcanic Wind's amount: the creatures
// on the battlefield as the spell is cast (the division is announced with the
// targets, CR 601.2d), so a creature that dies in response changes nothing.
var OldB690DivideCreaturesOnBattlefield = game.RegisterDivideAmount("creatures-on-the-battlefield",
	func(g *game.Game, _ game.DivideAmountArgs) game.DivideSpec {
		n := 0
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].IsCreature() {
				n++
			}
		}
		return game.DivideSpec{Total: n}
	})
