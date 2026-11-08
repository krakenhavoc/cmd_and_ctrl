package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archway Angel — Creature — Angel {5}{W}, 3/4 (Foundations, #4465):
//
//	"Flying
//	 When this creature enters, you gain 2 life for each Gate you
//	 control."
//
// Aven Gagglemaster's count with Gates in place of fliers: the Gates
// you control are counted as the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f2cf087a-d89a-4145-9394-945ce41c8cca",
		Name:            "Archway Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Archway Angel — you gain 2 life for each Gate you control", func(g *game.Game, item *game.StackItem) error {
				gates := permanentsControlledByMatching(g, item.Controller, HasSubtype("Gate"))
				return GainLife{Amount: 2 * len(gates)}.Apply(NewContext(g, item))
			}),
		},
	})
}
