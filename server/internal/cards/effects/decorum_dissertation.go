package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Decorum Dissertation — Sorcery — Lesson {3}{B}{B}:
//
//	"Target player draws two cards and loses 2 life.
//	 Paradigm (Then exile this spell. After you first resolve a spell
//	 with this name, you may cast a copy of it from exile without
//	 paying its mana cost at the beginning of each of your first main
//	 phases.)"
//
// The body is Sign in Blood's, word for word, so it shares that card's
// resolution. Paradigm is the printed keyword, and the engine does the
// rest (game/paradigm.go, #1866): the spell exiles itself as it
// resolves, and the first time you resolve one, each of your precombat
// main phases for the rest of the game offers you a free cast of a copy
// of it, which picks its own target player.
func init() {
	Register(Spec{
		OracleID:        "f4c57454-d36f-4f38-b7ac-76477a516c3e",
		Name:            "Decorum Dissertation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordParadigm},
		Targets:         TargetPlayer("target player"),
		Purpose:         ForTargets(game.TargetPurpose{Slot: 0, Draws: 2, LifeLoss: 2}),
		OnResolve:       signInBloodOnResolve,
	})
}
