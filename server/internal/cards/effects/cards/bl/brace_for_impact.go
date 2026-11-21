package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brace for Impact — Instant {4}{W}:
//
//	"Prevent all damage that would be dealt to target multicolored
//	 creature this turn. For each 1 damage prevented this way, put a
//	 +1/+1 counter on that creature."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the not-one-use shield pinned to
// the target, with a CR 615.5 follow-up. The follow-up runs once for each
// instance of damage the shield prevents, immediately after it, with
// that instance's total (CR 615.13's "each time a prevention effect is
// applied"), and puts that many counters on the creature. Damage that
// can't be prevented is dealt and gives no counters (CR 615.12).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f250fb94-e519-4bfa-921e-eddeeafcf76c",
		Name:         "Brace for Impact",
		Purpose:      game.Purpose{Answers: game.AnswerPump | game.AnswerPrevent}, // #4734 slice 6
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target multicolored creature", effects.Multicolored()),
		OnResolve: effects.SourceShieldSpell(effects.PreventDamageFromSource{
			Protect: effects.ShieldTheTarget,
			Then:    effects.CountersOnItPerPreventedBody,
		}),
	})
}
