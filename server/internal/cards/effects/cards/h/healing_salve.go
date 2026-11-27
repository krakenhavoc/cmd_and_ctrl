package cardsh

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Healing Salve — Instant {W} (Eighth Edition, #5607):
//
//	"Choose one —
//	 • Target player gains 3 life.
//	 • Prevent the next 3 damage that would be dealt to any target this turn."
//
// Healing Leaves' two modes (the same printed text), each with its own target clause chosen as the spell
// is cast (CR 700.2a): a player for the life gain, any target for the charged
// shield (Aurora Eidolon's, CR 615.7), which lasts until it is used up or the
// turn ends. A target that is gone by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8da8644c-75a1-4fe9-8e94-900d948d631c",
		Name:         "Healing Salve",
		Completeness: effects.CompletenessFull,
		Modes:        effects.OldB608GainThreeOrPreventThree("Healing Salve"),
	})
}
