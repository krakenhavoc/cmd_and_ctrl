package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moment of Heroism — Instant {1}{W} (Jumpstart Azorius deck request,
// #4894):
//
//	"Target creature gets +2/+2 and gains lifelink until end of turn."
//
// Both halves land on the creature while it is still a legal target
// (CR 608.2b); one that left in response gets neither.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a167d593-cded-43e1-b9bb-2287dbd0cbec",
		Name:         "Moment of Heroism",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2, "lifelink")),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			const label = "Moment of Heroism — +2/+2 and lifelink until end of turn"
			if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: label}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"lifelink"}, Label: label}.Apply(ctx)
		},
	})
}
