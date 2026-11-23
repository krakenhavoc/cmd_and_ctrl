package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lys Alana Informant — Creature — Elf Scout {1}{G} (Lorwyn Eclipsed, #4458):
//
//	"When this creature enters or dies, surveil 1."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5058b5c-0359-42ec-ba9e-ae1371bb9c22",
		Name:         "Lys Alana Informant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrDies("Lys Alana Informant — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
	})
}
