package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lightshell Duo — Creature — Rat Otter {3}{U}, 3/4 (Bloomburrow, #4467):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 When this creature enters, surveil 2."
//
// Prowess is a keyword the engine derives its trigger from
// (game/prowess.go); the entry surveil is Imperious Inkmage's (CR
// 701.25a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ef8fab12-d7d2-4c21-8b8c-41627048ac8b",
		Name:            "Lightshell Duo",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lightshell Duo — surveil 2", Do(Surveil{N: 2})),
		},
	})
}
