package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// std_b67_helpers.go — shared by the Lorwyn Eclipsed batches 67–70 (#4458).

// withAnswers is p with the answers an ability gives declared beside its
// target entries: ForTargets(...) on its own carries only the targets.
func withAnswers(p game.Purpose, a game.Answers) game.Purpose {
	p.Answers = a
	return p
}

// modeTargetPlayerDraws is the "Target player draws a card." bullet: the
// mode occurrence's target player draws one card, nothing if the target is
// no longer legal (Sublime Epiphany, Sygg's Command).
func modeTargetPlayerDraws(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return DrawCards{Player: t.ID, N: 1}.Apply(ctx)
}
