package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pym Particles — Sorcery {U} (Marvel Super Heroes, #4455):
//
//	"Target creature gains vigilance until end of turn and can't be blocked
//	 this turn.
//	 Draw a card."
//
// Taigam's Strike's shape: a grant and an unblockable restriction on the one
// target, both for the turn, then the draw. A target that is gone by
// resolution counters the spell by game rules (CR 608.2b), so the draw goes
// with it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "636b44dc-f7e4-42d2-abf6-ae49432b6052",
		Name:         "Pym Particles",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(CantBeBlockedUntilEOT(0, 0, 0, "vigilance")),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (GrantKeywordUntilEOT{Target: target, Keywords: []string{"vigilance"}, Label: "Pym Particles — vigilance"}).Apply(ctx); err != nil {
				return err
			}
			if err := (RestrictUntilEOT{Target: target, Restrictions: game.CantBeBlocked, Label: "Pym Particles — can't be blocked"}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
