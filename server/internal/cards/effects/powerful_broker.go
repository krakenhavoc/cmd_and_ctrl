package effects

import (
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Powerful Broker — Creature — Human Villain {2}{G}, 3/3 (Marvel Super
// Heroes, #4455):
//
//	"{T}: For each kind of counter on target permanent or player, give
//	 that permanent or player another counter of that kind. Activate only
//	 as a sorcery."
//
// Not proliferate: it is not the keyword action, so a "proliferate twice
// instead" replacement (Tekuthal) does not apply and the target is the
// activation's, not a choice at resolution. The kinds are read as the
// ability resolves (a permanent with +1/+1 and charge counters gets one of
// each; a player with poison and energy likewise) and snapshotted before
// the first lands, so a doubler on one kind does not change what the next
// kind receives. The target is Ardenn's "target permanent or player".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2b3d76e7-faaa-4e6f-9e21-a88dd719247e",
		Name:         "Powerful Broker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{T}: For each kind of counter on target permanent or player, give that permanent or player another counter of that kind. Activate only as a sorcery.",
			Cost:         TapCost(),
			SorcerySpeed: true,
			Targets:      ardennTargetPermanentOrPlayer(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return mshA6AnotherOfEachKind(ctx, t)
				}
				return nil
			},
		}},
	})
}

// mshA6AnotherOfEachKind gives the target one more counter of each kind it
// has now.
func mshA6AnotherOfEachKind(ctx *Context, t game.TargetRef) error {
	if t.Kind == game.TargetPlayer {
		p := ctx.PlayerByID(t.ID)
		if p == nil {
			return nil
		}
		kinds := mshA6KindsWithCounters(p.Counters)
		for _, kind := range kinds {
			if err := ctx.Game.AddPlayerCounterForEffect(t.ID, kind, 1); err != nil {
				return err
			}
		}
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(t.ID)
	if !ok {
		return nil
	}
	for _, kind := range mshA6KindsWithCounters(c.Counters) {
		if err := (AddCounter{Target: t.ID, Kind: kind, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// mshA6KindsWithCounters lists, in a stable order, the kinds with at least
// one counter.
func mshA6KindsWithCounters(counters map[string]int) []string {
	kinds := make([]string, 0, len(counters))
	for kind, n := range counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	return kinds
}
