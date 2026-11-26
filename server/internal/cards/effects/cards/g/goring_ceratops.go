package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goring Ceratops — Creature — Dinosaur {5}{W}{W}, 3/3 (Ixalan, S82 older
// sets):
//
//	"Double strike
//	 Whenever this creature attacks, other creatures you control gain double
//	 strike until end of turn."
//
// The set of creatures is locked as the ability resolves (CR 611.2c), as
// Loyal Unicorn's grant is: creatures that arrive later this turn do not gain
// double strike.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "34f5383b-21fe-45f5-9598-7382ee4c5b12",
		Name:            "Goring Ceratops",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Goring Ceratops — other creatures you control gain double strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID)),
						Keywords: []string{"double strike"},
						Label:    "Goring Ceratops — other creatures you control gain double strike",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
