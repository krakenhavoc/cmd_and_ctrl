package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aven Brigadier — Creature — Bird Soldier {3}{W}{W}{W}, 3/5
// (Onslaught):
//
//	"Flying
//	 Other Bird creatures get +1/+1.
//	 Other Soldier creatures get +1/+1."
//
// Two lords, Lord of the Undead's shape twice. Neither says "you
// control", so every other Bird and every other Soldier at the table
// grows, an opponent's included. They are two separate abilities, so a
// creature that is both a Bird and a Soldier gets +2/+2. The Brigadier
// never pumps itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "910ae0ca-257d-4b45-b039-f26e7b2f3d5c",
		Name:            "Aven Brigadier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Bird"}, Others: true}, 1, 1),
			TribalAnthem(TribeFilter{Tribes: []string{"Soldier"}, Others: true}, 1, 1),
		},
	})
}
