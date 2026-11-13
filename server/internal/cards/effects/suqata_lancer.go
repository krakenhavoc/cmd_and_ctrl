package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Suq'Ata Lancer — Creature — Human Knight {2}{R}, 2/2 (Visions, #5631):
//
//	"Haste
//	 Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)"
//
// Haste is an engine keyword carried as printed data (CR 702.10b).
// Flanking is the shared trigger in flanking.go (CR 702.25a), one per
// blocker without flanking.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f348b326-8a73-4df9-8faf-e271a36ce373",
		Name:            "Suq'Ata Lancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste", game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{Flanking("Suq'Ata Lancer")},
	})
}
