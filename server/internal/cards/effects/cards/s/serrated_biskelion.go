package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Serrated Biskelion — Artifact Creature — Construct {3}, 2/2 (Duel Decks:
// Elspeth vs. Tezzeret, #5578):
//
//	"{T}: Put a -1/-1 counter on this creature and a -1/-1 counter on target
//	 creature."
//
// The target is any creature, the Biskelion itself included. The two counters
// go on together as the ability resolves (CR 608.2): the Biskelion's only
// while it is still the same permanent on the battlefield (CR 400.7), the
// target's only while it is still legal (CR 608.2b); either half stands
// without the other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8552706-83f2-41d7-8068-75891abb8bde",
		Name:         "Serrated Biskelion",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Put a -1/-1 counter on this creature and a -1/-1 counter on target creature.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  S82OldB474BiskelionCounters,
		}},
	})
}

// S82OldB474BiskelionCounters puts a -1/-1 counter on the source (if it is
// still on the battlefield as the same object) and one on the first target
// still legal.
func S82OldB474BiskelionCounters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if effects.OnBattlefield(g, item.SourceCardID) && !effects.SourceIsNewObject(g, item) {
		if err := effects.NothingIfGone((effects.AddCounter{Target: item.SourceCardID, Kind: "-1/-1", N: 1}).Apply(ctx)); err != nil {
			return err
		}
	}
	if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
		return effects.NothingIfGone((effects.AddCounter{Target: id, Kind: "-1/-1", N: 1}).Apply(ctx))
	}
	return nil
}
