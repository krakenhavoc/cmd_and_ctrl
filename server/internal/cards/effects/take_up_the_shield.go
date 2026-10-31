package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Take Up the Shield — {1}{W} Instant:
//
//	"Put a +1/+1 counter on target creature. It gains lifelink and
//	 indestructible until end of turn."
//
// The counter is permanent; the keywords last until end of turn
// (CR 611.2a). If the target is gone as the spell resolves, nothing happens
// (CR 608.2b).
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	Register(Spec{
		OracleID:     "e8d0eafe-540a-4b2e-a989-e98ff3c31105",
		Name:         "Take Up the Shield",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "lifelink", "indestructible")),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"lifelink", "indestructible"},
					Label:    "Take Up the Shield — lifelink and indestructible",
				}.Apply(ctx)
			}
			return nil
		},
	})
}
