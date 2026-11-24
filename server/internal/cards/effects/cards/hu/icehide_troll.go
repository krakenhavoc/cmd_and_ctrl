package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Icehide Troll — Snow Creature — Troll Warrior {2}{G}, 2/3:
//
//	"{S}{S}: This creature gets +2/+0 and gains indestructible until end of
//	 turn. Tap it. ({S} can be paid with one mana from a snow source.)"
//
// Hailstorm Valkyrie's {S}{S} pump with Bloodsworn Squire's tail: the pump and
// the keyword are until-end-of-turn effects on the Troll and the tap follows
// them, so the Troll is tapped but already indestructible and bigger. A Troll
// that has left does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce55c9a8-25ba-404c-95b4-c14ba4763081",
		Name:         "Icehide Troll",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{S}{S}: This creature gets +2/+0 and gains indestructible until end of turn. Tap it.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2}},
			Cost:    effects.ManaCost("{S}{S}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.ThisGetsUntilEndOfTurn(2, 0, "Icehide Troll — +2/+0")(g, item); err != nil {
					return err
				}
				if err := effects.ThisGainsKeywordUntilEndOfTurn("indestructible", "Icehide Troll — indestructible until end of turn")(g, item); err != nil {
					return err
				}
				return effects.TapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
