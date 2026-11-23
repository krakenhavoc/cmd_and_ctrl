package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stand Up for Yourself — Instant {2}{W} (Secrets of Strixhaven, #4456):
//
//	"Destroy target creature with power 3 or greater."
//
// Bring Down's shape: the pick is a creature with power 3 or greater
// (PowerGE), re-checked at resolution (CR 608.2b), so a creature shrunk
// in response is an illegal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00e94b75-ad31-4e1f-b3be-eed7cd3e4d76",
		Name:         "Stand Up for Yourself",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with power 3 or greater", effects.PowerGE(3)),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
