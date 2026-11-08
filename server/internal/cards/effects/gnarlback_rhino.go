package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gnarlback Rhino — Creature — Rhino {2}{G}{G}, 4/4 (Foundations, #4465):
//
//	"Trample
//	 Whenever you cast a spell that targets this creature, draw a card."
//
// Favored Hoplite's cast trigger (heroic, without the ability word) with a
// draw. The spell's targets are on its stack item as it becomes cast (CR
// 601.2c), so only your own spells count, a spell that targets the Rhino
// twice draws once, and a copy (not cast, CR 707.10) or an ability draws
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1ae339fd-8e64-4e3a-8e27-e4993932ba62",
		Name:            "Gnarlback Rhino",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			heroic("Gnarlback Rhino — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
