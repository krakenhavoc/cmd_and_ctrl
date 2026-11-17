package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// bro01AmbushParatrooperPump is Ambush Paratrooper's "{5}: Creatures you
// control get +1/+1 until end of turn." (CR 611.2c: the set of creatures
// is fixed when the ability resolves.)
func bro01AmbushParatrooperPump() ActivatedAbility {
	return ActivatedAbility{
		Label: "{5}: Creatures you control get +1/+1 until end of turn.",
		Purpose: game.Purpose{
			Answers:  game.AnswerPump,
			TeamPump: &game.Pump{Power: 1, Toughness: 1},
		},
		Cost: ManaCost("{5}"),
		Effect: func(g *game.Game, item *game.StackItem) error {
			return BoostUntilEOT{
				Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
				Label: "Ambush Paratrooper — creatures you control get +1/+1",
			}.Apply(NewContext(g, item))
		},
	}
}
