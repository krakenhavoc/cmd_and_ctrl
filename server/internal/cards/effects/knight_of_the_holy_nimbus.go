package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knight of the Holy Nimbus — Creature — Human Rebel Knight {W}{W}, 2/2:
//
//	"Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)
//	 If this creature would be destroyed, regenerate it.
//	 {2}: This creature can't be regenerated this turn. Only your
//	 opponents may activate this ability."
//
// Clergy of the Holy Nimbus's regeneration and opponents-only row
// (noRegenerationRow) at {2}, plus flanking: the keyword token, which the
// engine triggers from (game/flanking.go, #5784). The row beside it is the
// retired pre-#5784 trigger, kept only for restore (RetiredFlankingRow).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "4e92c705-19c2-42df-ad12-808d272c505e",
		Name:            "Knight of the Holy Nimbus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{RetiredFlankingRow("Knight of the Holy Nimbus")},
		Replacements:    []game.ReplacementEffect{RegenerateIfThisWouldBeDestroyed()},
		Activated:       []ActivatedAbility{noRegenerationRow("{2}")},
	})
}
