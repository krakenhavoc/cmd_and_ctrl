package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82Bns18SourcesDamageTheOther is Band Together's resolution: slot 0's
// creatures that are still legal each deal damage equal to their power to
// slot 1's creature, as one damage event (eachDealsDamageEqualToItsPower).
// With no source chosen, or a victim that has left, nothing happens
// (CR 608.2b).
func S82Bns18SourcesDamageTheOther(ctx *Context) error {
	var sources []game.TargetRef
	for _, t := range ctx.ClauseTargets(0) {
		if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
			sources = append(sources, t)
		}
	}
	victim, ok := ctx.ClauseTarget(1)
	if !ok || victim.Kind != game.TargetCard || len(sources) == 0 {
		return nil
	}
	return EachDealsDamageEqualToItsPower(ctx, sources, victim.ID)
}
