package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82DmuB01TeamPumpRow is "<cost>: Creatures you control get +1/+1 until
// end of turn." (Charismatic Vanguard).
func S82DmuB01TeamPumpRow(name, cost string) ActivatedAbility {
	return ActivatedAbility{
		Label: cost + ": Creatures you control get +1/+1 until end of turn.",
		Purpose: game.Purpose{
			Answers:  game.AnswerPump,
			TeamPump: &game.Pump{Power: 1, Toughness: 1},
		},
		Cost: ManaCost(cost),
		Effect: func(g *game.Game, item *game.StackItem) error {
			return BoostUntilEOT{
				Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
				Label: name + " — creatures you control get +1/+1",
			}.Apply(NewContext(g, item))
		},
	}
}
