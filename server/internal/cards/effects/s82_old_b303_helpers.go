package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b303_helpers.go — shared pieces of Commander 2015 batch 303
// (S82 OLD #5546).

// ThreatenTheTarget is "Gain control of target creature until end of turn.
// Untap that creature. It gains haste until end of turn." as a spell's whole
// resolution (Act of Aggression), the three clauses in printed order. The
// theft is a layer-2 effect that ends at cleanup by itself, locked as the
// spell resolves (CR 611.2c). A target that is gone does nothing (CR 608.2b).
func ThreatenTheTarget(name string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		steps := []Applier{
			GainControl{Target: id, Duration: DurationUntilEndOfTurn(ctx), Label: name + " — gain control until end of turn"},
			UntapTarget{Target: id},
			GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: name + " — haste until end of turn"},
		}
		for _, s := range steps {
			if err := s.Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}
