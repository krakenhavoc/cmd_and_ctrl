package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Pegasus — Creature — Phyrexian Pegasus {2}{W}, 2/2 (March of
// the Machine):
//
//	"Flying
//	 Whenever this creature attacks, another target attacking creature
//	 without flying gains flying until end of turn."
//
// The target is chosen as the trigger goes on the stack and re-checked as
// it resolves (CR 608.2b); a creature that gained flying, stopped
// attacking or left in response gets nothing.
//
// No simplification.
func init() {
	const label = "Phyrexian Pegasus — another target attacking creature without flying gains flying until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "30236ff8-f6d5-462d-9a2f-846b2d8a1ca6",
		Name:            "Phyrexian Pegasus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks(label, effects.FirstTargetGainsUntilEndOfTurn(0, label, "flying")),
				effects.Another(effects.TargetCreature("another target attacking creature without flying",
					effects.AttackingCreature(), effects.WithoutKeyword("flying")))),
		},
	})
}
