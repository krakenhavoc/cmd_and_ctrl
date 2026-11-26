package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Duergar Mine-Captain — Creature — Dwarf Soldier {2}{R/W} 2/1:
//
//	"{1}{R/W}, {Q}: Attacking creatures get +1/+0 until end of turn. ({Q} is
//	 the untap symbol.)"
//
// {Q} is the untap-self cost (Umbral Mantle's), so the Captain must be tapped
// to activate; summoning sickness does not apply to it (CR 302.6). The
// attacking creatures as it resolves, any player's, are the ones pumped
// (CR 611.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ac276ca-91d5-4c7d-a355-e56c67ac3ba9",
		Name:         "Duergar Mine-Captain",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R/W}, {Q}: Attacking creatures get +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    game.AbilityCost{Mana: "{1}{R/W}", UntapSelf: true},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.B41Attacking()), Power: 1,
					Label: "Duergar Mine-Captain — attacking creatures +1/+0 until end of turn"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
