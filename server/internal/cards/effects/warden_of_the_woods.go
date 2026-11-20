package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warden of the Woods — Creature — Treefolk {4}{G}{G}, 5/7 (Core Set 2021, #5511):
//
//	"Vigilance
//	 Whenever this creature becomes the target of a spell or ability an
//	 opponent controls, you may draw two cards."
//
// Cactarantula's trigger (blackBoltTargetedByAnOpponent: the event names
// the Warden and its actor is someone else) drawing two, with the "may"
// asked as the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c8c6924e-9af8-4212-be28-b0b4eaa4c7d8",
		Name:            "Warden of the Woods",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			OptionalDraw(On(game.EventBecomesTarget, blackBoltTargetedByAnOpponent,
				"Warden of the Woods — you may draw two cards", Do(DrawCards{N: 2})),
				"Warden of the Woods — draw two cards?", 2),
		},
	})
}
