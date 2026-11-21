package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beverly Crusher, Ship's Doctor — Legendary Creature — Human Doctor
// {2}{W}, 2/2 (Star Trek Commander, #4974):
//
//	"Seek out — Whenever you draw your second card each turn, put a
//	 +1/+1 counter on each of up to three target creatures."
//
// The seek-out trigger (trekB31SeekOut) with Rishkar's resolution: zero
// to three creatures, any controller, chosen as the trigger goes on the
// stack; only those still legal as it resolves get a counter (CR
// 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a33edc3-e938-463f-8568-7d7a458427c8",
		Name:         "Beverly Crusher, Ship's Doctor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.TrekB31SeekOut("Beverly Crusher — a +1/+1 counter on each of up to three target creatures",
					effects.PutPlusOneCounterOnEachLegalTarget),
				effects.TargetCreature("up to three target creatures").WithCount(0, 3)),
		},
	})
}
