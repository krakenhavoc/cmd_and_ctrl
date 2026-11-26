package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Groundling Pouncer — Creature — Faerie {1}{G/U} 2/1:
//
//	"{G/U}: This creature gets +1/+3 and gains flying until end of turn.
//	 Activate only once each turn and only if an opponent controls a creature
//	 with flying."
//
// Basking Rootwalla's once-per-turn gate (it counts activations announced,
// CR 602.5b) joined to Night Revelers' "an opponent controls a ..." check
// (one opponent, not summed); both are read as the ability is announced.
func init() {
	const label = "{G/U}: This creature gets +1/+3 and gains flying until end of turn. Activate only once each turn and only if an opponent controls a creature with flying."
	effects.Register(effects.Spec{
		OracleID:     "f695a073-7235-45c8-bc42-f1e29bb086ee",
		Name:         "Groundling Pouncer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   label,
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 3}},
			Cost:    effects.ManaCost("{G/U}"),
			Condition: effects.AllConditions(
				effects.OncePerTurnActivation(label),
				effects.OpponentControlsAtLeast(1, func(c game.Card) bool { return game.HasKeyword(&c, "flying") }),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.ThisGetsUntilEndOfTurn(1, 3, "Groundling Pouncer — +1/+3 until end of turn")(g, item); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"flying"}, Label: "Groundling Pouncer — gains flying until end of turn"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
