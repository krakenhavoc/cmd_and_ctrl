package effects

import (
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
const orochiEggwatcherOracleID = "e448e27c-f66d-4369-8215-7f7275e456e1"

func init() {
	Register(Spec{
		OracleID:     orochiEggwatcherOracleID,
		Name:         "Orochi Eggwatcher",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}, {T}: Create a 1/1 green Snake creature token. If you control ten or more creatures, flip this creature.",
			Cost:    Plus(ManaCost("{2}{G}"), TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := (CreateToken{Template: TokenCard("1/1 green Snake"), N: 1}).Apply(NewContext(g, item)); err != nil {
					return err
				}
				return flipThisWhen(g, item, controlsAtLeast(10, game.Card.IsCreature))
			},
		}},
	})
	Register(Spec{
		OracleID:     orochiEggwatcherOracleID + "#1",
		Name:         "Shidako, Broodmistress",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{G}, Sacrifice a creature: Target creature gets +3/+3 until end of turn.",
			Cost:    Plus(ManaCost("{G}"), SacrificeACreature()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if id, ok := b16FirstLegalTargetCard(ctx); ok {
					return BoostUntilEOT{Target: id, Power: 3, Toughness: 3, Label: "Shidako, Broodmistress — +3/+3"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
