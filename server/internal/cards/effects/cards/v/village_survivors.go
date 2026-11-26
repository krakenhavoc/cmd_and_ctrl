package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Village Survivors — Creature — Human {4}{G} 4/5:
//
//	"Vigilance
//	 Fateful hour — As long as you have 5 or less life, other creatures
//	 you control have vigilance."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c26dd36f-a81a-4d97-9be2-d350e93da79e",
		Name:            "Village Survivors",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static:          []game.StaticAbility{effects.OldB441OtherCreaturesKeywordWhileFatefulHour("vigilance")},
	})
}
