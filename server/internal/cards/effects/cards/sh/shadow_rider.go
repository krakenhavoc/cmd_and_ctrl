package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadow Rider — Creature — Knight {2}{B}{B}, 3/3 (Weatherlight, #5629):
//
//	"Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)"
//
// Flanking is the shared trigger in flanking.go (CR 702.25a), one per
// blocker without flanking, as on Fallen Askari.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a518b31a-7e9a-460d-839f-39cb1de4f570",
		Name:            "Shadow Rider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{effects.RetiredFlankingRow("Shadow Rider")},
	})
}
