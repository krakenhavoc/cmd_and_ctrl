package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daybreak Chaplain — Creature — Human Cleric {1}{W}, 1/3:
//
//	"Lifelink"
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ebd33f4c-2cc8-46e3-b367-ce5b37c2719b",
		Name:            "Daybreak Chaplain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
