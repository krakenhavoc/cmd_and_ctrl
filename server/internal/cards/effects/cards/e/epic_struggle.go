package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Epic Struggle — Enchantment {2}{G}{G} (Judgment):
//
//	"At the beginning of your upkeep, if you control twenty or more
//	 creatures, you win the game."
//
// Biovisionary's win-the-game trigger with an intervening "if" (CR 603.4):
// checked as it triggers and again as it resolves, on the controller's own
// upkeep only. Tokens count.
//
// No simplification.
func init() {
	const name = "Epic Struggle"
	t := effects.AtYourUpkeep(name+" — you win the game", effects.S82OldB626YouWinTheGame)
	effects.Register(effects.Spec{
		OracleID:     "6f34ee2c-6e0b-4c71-9bdd-6ac55a323849",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control twenty or more creatures", func(g *game.Game, you uuid.UUID) bool {
				n := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == you && c.IsCreature() {
						n++
					}
				}
				return n >= 20
			}, t),
		},
	})
}
