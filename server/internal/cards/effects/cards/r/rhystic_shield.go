package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rhystic Shield — Instant {1}{W}:
//
//	"Creatures you control get +0/+1 until end of turn. They get an
//	 additional +0/+2 until end of turn unless any player pays {2}."
//
// ADR 0150 (#1903): the +0/+1 lands as the spell resolves; the +0/+2
// waits on the {2} offered to every player in APNAP order from the
// active player (AnyPlayerPays). The offer holds the step and nobody
// gets priority in between, so "they" are the same creatures (CR
// 611.2c); one that left meanwhile is not there to get it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a905dafa-1818-471a-a485-d84be447263b",
		Name:         "Rhystic Shield",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Toughness: 1, Label: "Rhystic Shield — +0/+1"}).Apply(ctx); err != nil {
				return err
			}
			return effects.AnyPlayerPays{
				Cost:     "{2}",
				Question: "Rhystic Shield — pay {2} to stop the additional +0/+2?",
				Then:     rhysticShieldExtra,
			}.Apply(ctx)
		},
	})
}

// rhysticShieldExtra is "if no one does", keyed (#4264).
var rhysticShieldExtra = effects.AnyPlayerConsequences("any-player/rhystic-shield-extra",
	nil,
	func(ctx *effects.Context, _ game.AnyPlayerAnswer) error {
		return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Toughness: 2, Label: "Rhystic Shield — an additional +0/+2"}.Apply(ctx)
	},
	nil)
