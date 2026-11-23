package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knight of the Last Breath — Creature — Giant Knight {5}{W}{B}, 4/4 (Magic
// Commander Collection, #5908):
//
//	"{3}, Sacrifice another nontoken creature: Create a 1/1 white and black
//	 Spirit creature token with flying.
//	 Afterlife 3 (When this creature dies, create three 1/1 white and black
//	 Spirit creature tokens with flying.)"
//
// Afterlife is the dies trigger it reminds (CR 702.135a), Tithe Taker's
// shape with three tokens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "67791323-bf21-4c72-bf57-32197d3d5775",
		Name:         "Knight of the Last Breath",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, Sacrifice another nontoken creature: Create a 1/1 white and black Spirit creature token with flying.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: Plus(ManaCost("{3}"),
				SacrificeAnotherN(1, "another nontoken creature", Creature(), Not(IsTokenPredicate()))),
			Effect: KnightOfTheLastBreathSpirits(1),
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Knight of the Last Breath — afterlife 3: create three 1/1 white and black Spirit creature tokens with flying",
				KnightOfTheLastBreathSpirits(3)),
		},
	})
}

// knightOfTheLastBreathSpirits makes n of the printed 1/1 white and black
// Spirit tokens with flying. The row is added by std_b43_helpers.go's init,
// which runs after this file's, so the template is read as the effect runs.
func KnightOfTheLastBreathSpirits(n int) Effect {
	return Do(Instruction(func(ctx *Context) error {
		return CreateToken{Template: TokenCard("1/1 white and black Spirit with flying"), N: n}.Apply(ctx)
	}))
}
