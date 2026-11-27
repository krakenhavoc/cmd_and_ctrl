package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soulbound Guardians — {4}{W} Creature — Kor Spirit 4/5 (Rise of the Eldrazi, S82 OLD #5581):
//
//	Defender, flying

func init() {
	effects.Register(effects.Spec{
		OracleID:        "11f7f6e4-2c39-4478-8c10-4473dbf921a2",
		Name:            "Soulbound Guardians",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
	})
}
