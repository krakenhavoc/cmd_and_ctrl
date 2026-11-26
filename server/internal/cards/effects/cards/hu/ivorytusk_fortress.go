package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ivorytusk Fortress — Khans of Tarkir (S82 OLD b346, #5554):
//
//	Creature — Elephant {2}{W}{B}{G}, 5/7: Untap each creature you control with a
//	+1/+1 counter on it during each other player's untap step.
//
// The shared untap-step permission (Seedborn Muse), narrowed to creatures
// that carry a +1/+1 counter as the untap step begins.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10ae7744-5f98-4d6f-993a-ffb2a34b23e7",
		Name:         "Ivorytusk Fortress",
		Completeness: effects.CompletenessFull,
		UntapStep: []game.UntapStepPermission{
			effects.UntapDuringEachOtherPlayersUntapStep(
				"Ivorytusk Fortress — untap each creature you control with a +1/+1 counter on it",
				func(c game.Card) bool { return c.IsCreature() && c.Counters[game.CounterPlusOne] > 0 }),
		},
	})
}
