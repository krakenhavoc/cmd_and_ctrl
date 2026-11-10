package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valley Rotcaller — Creature — Squirrel Warlock {1}{B}, 1/3
// (Bloomburrow, #4467):
//
//	"Menace
//	 Whenever this creature attacks, each opponent loses X life and you
//	 gain X life, where X is the number of other Squirrels, Bats,
//	 Lizards, and Rats you control."
//
// Menace is a printed keyword. X is counted as the trigger resolves
// (CR 608.2h), over the permanents you control with any of the four
// subtypes other than this object, the subtypes read through the
// Rotcaller's text. Each opponent then loses X and you gain X once
// (Dai Li Agents' drain). With X = 0 nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8da0599a-8f63-4dba-857f-b4e10b7191b6",
		Name:            "Valley Rotcaller",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Valley Rotcaller — each opponent loses X life and you gain X life (X = other Squirrels, Bats, Lizards, and Rats you control)",
				valleyRotcallerDrain),
		},
	})
}

// valleyRotcallerDrain counts the other Squirrels, Bats, Lizards and Rats
// you control and drains each opponent for that many.
func valleyRotcallerDrain(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := len(MatchingBattlefield(ctx, And(
		AnySubtype("Squirrel", "Bat", "Lizard", "Rat"), YouControl(), OtherThan(item.SourceCardID))))
	if x <= 0 {
		return nil
	}
	return b40DrainEachOpponent(x)(g, item)
}
