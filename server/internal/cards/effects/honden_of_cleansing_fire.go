package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Honden of Cleansing Fire — Legendary Enchantment — Shrine {3}{W}
// (Champions of Kamigawa, #5603):
//
//	"At the beginning of your upkeep, you gain 2 life for each Shrine
//	 you control."
//
// The Shrines are counted as the trigger resolves (CR 608.2h), and the
// life is one gain of twice that.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "15ee7e1d-b13e-4fd0-bc41-345b6ed34f53",
		Name:         "Honden of Cleansing Fire",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			chkB1Honden("Honden of Cleansing Fire — gain 2 life for each Shrine you control",
				func(g *game.Game, item *game.StackItem, shrines int) error {
					return GainLife{Amount: 2 * shrines}.Apply(NewContext(g, item))
				}),
		},
	})
}
