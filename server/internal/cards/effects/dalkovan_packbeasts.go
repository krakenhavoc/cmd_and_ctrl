package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dalkovan Packbeasts — Creature — Ox {2}{W}, 0/4 (Tarkir: Dragonstorm,
// #4463):
//
//	"Vigilance
//	 Mobilize 3 (Whenever this creature attacks, create three tapped and
//	 attacking 1/1 red Warrior creature tokens. Sacrifice them at the
//	 beginning of the next end step.)"
//
// Mobilize is mobilize.go (CR 702.181a), Shock Brigade's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cdd1a043-bd47-4197-be26-e2f7d3f1d209",
		Name:            "Dalkovan Packbeasts",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{Mobilize("Dalkovan Packbeasts", 3)},
	})
}
