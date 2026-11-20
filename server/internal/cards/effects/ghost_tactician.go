package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghost Tactician — Creature — Spirit Spellshaper {4}{W}, 2/5 (Planar Chaos, #5595):
//
//	"{W}, {T}, Discard a card: Creatures you control get +1/+0 until end of
//	 turn."
//
// Castle Embereth's team pump with a discard in the cost: the creatures
// are read as the ability resolves, so one that arrives later is not
// pumped (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8d89aba8-20ce-44cd-9d82-3d16fdf6aac8",
		Name:         "Ghost Tactician",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}, Discard a card: Creatures you control get +1/+0 until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1}, Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{W}"), TapCost(), DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{
					Match: And(Creature(), YouControl()), Power: 1,
					Label: "Ghost Tactician — creatures you control get +1/+0",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
