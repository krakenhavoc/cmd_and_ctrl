package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thwip! — Instant {W} (Marvel Super Heroes Commander, #4455): "Target
// creature gets +2/+2 and gains flying until end of turn. If it's a
// Spider, you gain 2 life." The pump and the keyword are Smashing
// Spree's two steps; the life gain reads the target's subtype as the
// spell resolves (it is a Spider if it is one now, CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f7d11254-b48a-4016-a035-c0d6642fdfae",
		Name:         "Thwip!",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2, "flying")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Thwip! — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Thwip! — flying"}).Apply(ctx); err != nil {
				return err
			}
			for _, c := range MatchingBattlefield(ctx, Subtype("Spider")) {
				if c.InstanceID == id {
					return GainLife{Amount: 2}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
