package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7a33edc3-e938-463f-8568-7d7a458427c8",
		Name:         "Beverly Crusher, Ship's Doctor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				trekB31SeekOut("Beverly Crusher — a +1/+1 counter on each of up to three target creatures",
					putPlusOneCounterOnEachLegalTarget),
				TargetCreature("up to three target creatures").WithCount(0, 3)),
		},
	})
}
