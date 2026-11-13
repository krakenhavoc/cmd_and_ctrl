package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mtenda Herder — Creature — Human Scout {W}, 1/1 (Mirage, #5633):
//
//	"Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)"
//
// Flanking is flanking.go's trigger with the declared keyword (#5757,
// CR 702.25a). The -1/-1 resolves before combat damage, so a 1/1 blocker
// without flanking dies as a state-based action, and the Herder stays
// blocked with no creature to deal its damage to (CR 509.1h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0edf5bac-36a9-4f35-afd1-73f7e1aab51b",
		Name:            "Mtenda Herder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{Flanking("Mtenda Herder")},
	})
}
