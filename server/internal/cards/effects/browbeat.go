package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Browbeat — Sorcery {2}{R}:
//
//	"Any player may have Browbeat deal 5 damage to them. If no one
//	 does, target player draws three cards."
//
// ADR 0150 §3 (#1903): taking the damage is offered to every player in
// APNAP order from the active player, the caster included, until one
// takes it (anyPlayerMayTakeDamageFromThis). Browbeat deals it through
// the ordinary damage path, so it can be prevented, and taking it still
// stops the draw. The target is chosen as the spell is cast (CR
// 601.2c); a target gone by the time nobody pays draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11cd0c42-4148-4d82-9172-b1a1967fe319",
		Name:         "Browbeat",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve:    anyPlayerMayTakeDamageFromThis("Browbeat", 5, "the draw", browbeatDraws),
	})
}

// browbeatDraws is "If no one does, that player draws three cards", keyed
// (#4264): the spell's target rides the asking item.
var browbeatDraws = AnyPlayerConsequences("any-player/browbeat-draw", nil,
	func(ctx *Context, _ game.AnyPlayerAnswer) error {
		who, ok := firstLegalPlayerTarget(ctx)
		if !ok {
			return nil
		}
		return DrawCards{Player: who, N: 3}.Apply(ctx)
	}, nil)
