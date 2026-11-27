package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Etchings — Enchantment {B}{B}{B} (Coldsnap, S82 OLD b559):
//
//	"Cumulative upkeep {B}
//	 At the beginning of your end step, draw a card for each age counter on
//	 this enchantment.
//	 When this enchantment is put into a graveyard from the battlefield, you
//	 lose 2 life for each age counter on it."
//
// The draw counts the age counters as the end step trigger resolves; the life
// loss counts the ones it had as it left (CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf938a60-23ec-4834-a3cf-391a7b04747e",
		Name:         "Phyrexian Etchings",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Phyrexian Etchings — cumulative upkeep {B}", "{B}"),
			effects.AtYourEndStep("Phyrexian Etchings — draw a card for each age counter on it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DrawCards{Player: item.Controller, N: effects.OldB559AgeCounters(g, item)}.Apply(effects.NewContext(g, item))
				}),
			effects.WhenThisDies("Phyrexian Etchings — you lose 2 life for each age counter on it",
				func(g *game.Game, item *game.StackItem) error {
					n := effects.OldB559AgeCounters(g, item)
					if n <= 0 {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -2*n)
				}),
		},
	})
}
