package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fog Elemental — Creature — Elemental {2}{U}, 4/4 (Weatherlight, S82 older sets batch 780, #5629):
//
//	"Flying
//	 When this creature attacks or blocks, sacrifice it at end of combat."
//
// Flying rides PrintedKeywords. One trigger as it is declared as an attacker or as a blocker (a creature never does both in
// one combat), which schedules the sacrifice for the beginning of that combat's end of combat step (CR 511.2).
//
// No simplification.
func init() {
	const label = "Fog Elemental — sacrifice it at end of combat"
	effects.Register(effects.Spec{
		OracleID:        "b304fcc0-899e-433f-b6df-5cea504dc4ba",
		Name:            "Fog Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, effects.ThisAttacksOrBlocks,
				label, effects.SacrificeThisAtEndOfCombat(label)),
		},
	})
}
