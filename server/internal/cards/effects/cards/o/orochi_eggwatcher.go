package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Orochi Eggwatcher // Shidako, Broodmistress — {2}{G} Creature — Snake
// Shaman 1/1 // Legendary Creature — Snake Shaman 3/3, a flip card
// (#3216, ADR 0151):
//
//	Top:     "{2}{G}, {T}: Create a 1/1 green Snake creature token. If
//	          you control ten or more creatures, flip this creature."
//	Flipped: "{G}, Sacrifice a creature: Target creature gets +3/+3
//	          until end of turn."
//
// The creatures are counted after the token has arrived. Shidako's
// sacrifice may be itself or any creature you control, paid at
// announce. It is green and {2}{G} (CR 710.1c).
//
// No simplification.
const OrochiEggwatcherOracleID = "e448e27c-f66d-4369-8215-7f7275e456e1"

func init() {
	effects.Register(effects.Spec{
		OracleID:     OrochiEggwatcherOracleID,
		Name:         "Orochi Eggwatcher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, {T}: Create a 1/1 green Snake creature token. If you control ten or more creatures, flip this creature.",
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := (effects.CreateToken{Template: effects.TokenCard("1/1 green Snake"), N: 1}).Apply(effects.NewContext(g, item)); err != nil {
					return err
				}
				return effects.FlipThisWhen(g, item, effects.YouControlAtLeast(10, game.Card.IsCreature))
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     OrochiEggwatcherOracleID + "#1",
		Name:         "Shidako, Broodmistress",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, Sacrifice a creature: Target creature gets +3/+3 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)),
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.SacrificeACreature()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
					return effects.BoostUntilEOT{Target: id, Power: 3, Toughness: 3, Label: "Shidako, Broodmistress — +3/+3"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
