package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lys Alana Informant — Creature — Elf Scout {1}{G} (Lorwyn Eclipsed, #4458):
//
//	"When this creature enters or dies, surveil 1."
func init() {
	Register(Spec{
		OracleID:     "b5058b5c-0359-42ec-ba9e-ae1371bb9c22",
		Name:         "Lys Alana Informant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrDies("Lys Alana Informant — surveil 1", Do(Surveil{N: 1})),
		},
	})
}
