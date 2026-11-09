package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Great Aerie — Plane (Planechase)
// (#4870):
//
//	"When you planeswalk to The Great Aerie and at the beginning of your
//	 upkeep, bolster 3.
//	 Whenever chaos ensues, choose up to one target creature you control and
//	 up to one target creature an opponent controls. Each of those creatures
//	 deals damage equal to its toughness to the other."
//
// Bolster chooses on resolution (CR 701.39a). The chaos damage is not a
// fight: each creature is the source of its own damage, and it needs both
// creatures, so with only one chosen nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "00f00001-bd84-4dbf-a707-f4b1549100d4",
		Name:         "The Great Aerie",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("The Great Aerie — bolster 3", Do(Bolster{N: 3})),
			Targeting(WheneverChaosEnsues("The Great Aerie — up to one creature you control and up to one creature an opponent controls each deal damage equal to their toughness to the other",
				b6ToughnessDamagePair),
				Clauses(
					TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1),
					TargetCreature("up to one target creature an opponent controls", OpponentControls()).WithCount(0, 1),
				)),
		},
	})
}
