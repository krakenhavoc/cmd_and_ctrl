package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kraken of the Straits — {5}{U}{U} Creature — Kraken 6/6 (Born of the Gods):
//
//	"Creatures with power less than the number of Islands you control can't block this
//	 creature."
//
// Aura Gnarlid's block rule: the blocker's power is compared with the number of Islands the
// Kraken's controller controls, both as they are when the block is declared.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c282581f-3d77-47ce-ac92-ed6217754e38",
		Name:         "Kraken of the Straits",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedByComparing(effects.OnSelf(), effects.PowerLessThanIslandsYouControl(),
				"creatures with power less than the number of Islands you control"),
		},
	})
}
