package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scurrilous Sentry — Creature — Human Knight Rogue {3}{B}, 2/3 (Modern
// Horizons 3, #5467):
//
//	"Menace
//	 Whenever this creature enters or attacks, it connives. (Draw a
//	 card, then discard a card. If you discarded a nonland card, put a
//	 +1/+1 counter on this creature.)"
//
// Atlantean Skirmisher's connive on Sun Titan's two conditions: one
// ability watching both its own entry and its own attack. The counter
// goes on the Sentry only while it is still there (CR 701.50c). Menace
// is the engine's keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "52bfcb09-007c-47d8-8038-692e7dd8f922",
		Name:            "Scurrilous Sentry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Scurrilous Sentry — it connives", Do(Instruction(sourceConnives))),
		},
	})
}
