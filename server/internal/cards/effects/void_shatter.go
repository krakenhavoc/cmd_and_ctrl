package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Void Shatter — Instant {1}{U}{U} (Oath of the Gatewatch, #5545):
//
//	"Devoid (This card has no color.)
//	 Counter target spell. If that spell is countered this way, exile
//	 it instead of putting it into its owner's graveyard."
//
// Devoid makes it colourless (CR 702.114a). Dissipate's body: a spell
// countered this way is exiled instead of going to its owner's
// graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3d9f854e-1ee5-4aa1-a33b-d3ae08f15dd0",
		Name:            "Void Shatter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         game.Purpose{Answers: game.AnswerRemove},
		Targets:         TargetSpell("target spell"),
		OnResolve:       counterTargetSpellAndExileIt,
	})
}
