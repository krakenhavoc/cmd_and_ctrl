package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wary Thespian — Creature — Cat Druid {1}{G}, 3/1 (Foundations, #4465):
//
//	"When this creature enters or dies, surveil 1."
//
// Thawbringer's enters-or-dies trigger on a smaller body: one ability
// with two conditions, so it surveils once as it arrives and once more
// as it dies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "448cc829-1596-4aec-b180-c0c1e0566f6b",
		Name:         "Wary Thespian",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrDies("Wary Thespian — surveil 1", Do(Surveil{N: 1})),
		},
	})
}
