package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood-Chin Rager — Creature — Human Warrior {1}{B}, 2/2 (Dragons of Tarkir):
//
//	"Whenever this creature attacks, Warrior creatures you control gain menace
//	 until end of turn."
//
// The Warriors are those its controller controls as the trigger resolves (CR
// 611.2c), itself included; a Warrior arriving later this turn gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7036d08-6ee4-49b1-bbbf-c96a937afa15",
		Name:         "Blood-Chin Rager",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Blood-Chin Rager — Warrior creatures you control gain menace until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.Creature(), effects.YouControl(), effects.OfSubtype("Warrior")),
						Keywords: []string{"menace"},
						Label:    "Blood-Chin Rager — menace until end of turn",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
