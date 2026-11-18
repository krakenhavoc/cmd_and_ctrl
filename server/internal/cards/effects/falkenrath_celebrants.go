package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Falkenrath Celebrants — Creature — Vampire {4}{R}, 4/4 (Innistrad: Crimson Vow):
//
//	"Menace
//	 When this creature enters, create two Blood tokens."
func init() {
	Register(Spec{
		OracleID:        "3c0be888-0d66-4bad-84f4-90c3934915d7",
		Name:            "Falkenrath Celebrants",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Falkenrath Celebrants — create two Blood tokens", Do(s82Vow03CreateBlood(2))),
		},
	})
}
