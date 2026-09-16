package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sylvan Reclamation — Instant for {3}{G}{W}:
//
//	"Exile up to two target artifacts and/or enchantments.
//	 Basic landcycling {2}"
//
// S20 sub-PR 5's "up to N" card: Min 0, so it can be cast with no
// targets at all (and then does nothing). Basic landcycling isn't
// modelled, so the card is castable only. It is an activated ability
// from hand (typecycling, CR 702.29e), not a cast, and waits on a
// discard cost and activation from hand (#660).
func init() {
	Register(Spec{
		OracleID: "aeec8e85-6571-4da6-8a48-f5d3985ca10b",
		Name:     "Sylvan Reclamation",
		Targets: TargetPermanent("up to two target artifacts and/or enchantments",
			Or(Artifact(), Enchantment())).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
