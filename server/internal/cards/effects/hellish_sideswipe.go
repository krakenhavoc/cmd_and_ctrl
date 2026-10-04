package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hellish Sideswipe — Sorcery {B}:
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature. Destroy target creature or Vehicle. If the sacrificed
//	 permanent was a Vehicle, draw a card."
//
// "Was a Vehicle" is read off the sacrificed permanent as it last
// existed on the battlefield (CR 608.2h), from the payment record (ADR
// 0113 §1). The draw happens only if the spell resolves: an illegal
// target means it does not (CR 608.2b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "d919d8e9-d1ba-42de-9884-12e38dca78ac",
		Name:           "Hellish Sideswipe",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("an artifact or creature", Or(Artifact(), Creature())),
		Targets:        TargetPermanent("target creature or Vehicle", Or(Creature(), HasSubtype("Vehicle"))),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if legal := ctx.LegalTargets(); len(legal) > 0 && legal[0].Kind == game.TargetCard {
				if err := (DestroyTarget{Target: legal[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			if !sacrificedHadSubtype(ctx, "Vehicle") {
				return nil
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
