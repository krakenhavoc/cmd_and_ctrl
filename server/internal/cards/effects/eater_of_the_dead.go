package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eater of the Dead — Creature — Horror {4}{B}, 3/4 (Mystery Booster 2):
//
//	"{0}: If this creature is tapped, exile target creature card from a
//	 graveyard and untap this creature."
//
// The target is chosen whether or not the Eater is tapped; the condition is
// read as the ability resolves, and when it is not tapped nothing happens
// (the exile and the untap are one "if"). A source that has left does
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c414ace-fd04-4fac-9379-24f5910f20e4",
		Name:         "Eater of the Dead",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{0}: If this creature is tapped, exile target creature card from a graveyard and untap this creature.",
			Cost:    ManaCost("{0}"),
			Targets: TargetCardInGraveyard("target creature card from a graveyard", Creature()),
			Effect:  eaterOfTheDeadResolve,
		}},
	})
}

func eaterOfTheDeadResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !src.Tapped || ctx.isNewSourceObject(item.SourceCardID) {
		return nil
	}
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return nil
	}
	if err := (ExileTarget{Target: ts[0].ID}).Apply(ctx); err != nil {
		return err
	}
	return UntapTarget{Target: item.SourceCardID}.Apply(ctx)
}
