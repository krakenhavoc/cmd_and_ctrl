package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// YouControlTwoCreaturesSharingAType passes when the caster controls at
// least two creatures that share a creature type — Synchronized
// Eviction's "if you control at least two creatures that share a
// creature type". Types are read through the layered view, and a
// changeling shares a type with every creature that has one (CR 702.73a,
// game.SharesCreatureType).
func YouControlTwoCreaturesSharingAType() CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Game == nil || q.Game.Battlefield == nil {
			return false
		}
		var mine []game.Card
		for _, c := range q.Game.Battlefield.Cards {
			if c.Controller == q.Controller && c.IsCreature() {
				mine = append(mine, c)
			}
		}
		for i := range mine {
			for j := i + 1; j < len(mine); j++ {
				if game.SharesCreatureType(&mine[i], &mine[j]) {
					return true
				}
			}
		}
		return false
	}
}

// PutTargetPermanentSecondFromTheTop is "put target nonland permanent
// into its owner's library second from the top" — Synchronized
// Eviction.
func PutTargetPermanentSecondFromTheTop(_ *game.StackItem, ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
		return nil
	}
	return PutIntoLibrary{Card: t.ID, Depth: 2}.Apply(ctx)
}
