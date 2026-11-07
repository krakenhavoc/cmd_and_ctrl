package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// skip_combat_phases.go — the card side of "target player skips all
// combat phases of their next turn" (S79, #5075; CR 500.11, 614.10).
//
// TargetPlayerSkipsCombatPhasesOfNextTurn writes a record
// (game.SkipCombatPhasesOfNextTurnForEffect) that the engine arms as
// that player's next turn really begins, and every combat phase of that
// turn is then skipped, one added later in the turn included. It never
// touches the turn in progress: cast during the target's own turn (a
// copy, or a controlled turn), it waits for the next one (CR 614.10).
// Two of them skip the combat phases of the player's next two turns
// (CR 614.10a).

// TargetPlayerSkipsCombatPhasesOfNextTurn is "target player skips all
// combat phases of their next turn" (False Peace; Empty City Ruse's
// "target opponent"), reading the player from the card's player target
// slot. A target that left the game in response is no longer legal (CR
// 608.2b), so the effect does nothing for it.
type TargetPlayerSkipsCombatPhasesOfNextTurn struct{}

// Apply implements Primitive. It writes a record and never fails.
func (TargetPlayerSkipsCombatPhasesOfNextTurn) Apply(ctx *Context) error {
	label := ""
	if ctx.Item != nil {
		label = ctx.Item.Label
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			ctx.Game.SkipCombatPhasesOfNextTurnForEffect(t.ID, ctx.Source(), label)
		}
	}
	return nil
}
