package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Awe Strike — Instant {W}:
//
//	"The next time target creature would deal damage this turn, prevent that damage. You gain life equal to the damage prevented this way."
//
// ADR 0107 §6 (#1860): the shield is against the target creature's next
// instance of damage, to anything (CR 615.8), and "You gain life equal to
// the damage prevented this way" is its additional effect, run
// immediately after the prevention (CR 615.5, owner decision 4).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "994bcde3-4a76-403d-a6e1-88609a13a99c",
		Name:         "Awe Strike",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.ShieldAgainstTargetCreature(effects.PreventedGainLifeBody, false),
	})
}
