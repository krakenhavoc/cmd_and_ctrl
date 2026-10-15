package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unnatural Growth — Enchantment {1}{G}{G}{G}{G} (EDHREC top 1,000, #3654):
//
//	"At the beginning of each combat, double the power and toughness of
//	 each creature you control until end of turn."
//
// Zopandrel's trigger on its own: the set of creatures is fixed as the
// trigger resolves (CR 611.2c) and each one is doubled by its power
// and toughness at that moment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7324abaa-48da-439d-9339-b0ea5eea612e",
		Name:         "Unnatural Growth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfEachCombat("Unnatural Growth — double the power and toughness of each creature you control",
				func(g *game.Game, item *game.StackItem) error {
					var mine []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() {
							mine = append(mine, c.InstanceID)
						}
					}
					return DoublePowerAndToughnessOfEachUntilEOT(NewContext(g, item), mine,
						"Unnatural Growth — double power and toughness")
				}),
		},
	})
}
