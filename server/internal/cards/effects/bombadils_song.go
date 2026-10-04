package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bombadil's Song — Instant {1}{G}:
//
//	"Target creature you control gets +1/+1 and gains hexproof until
//	 end of turn. The Ring tempts you."
//
// Ranger's Guile, then the tempt. A target that has become illegal
// makes the whole spell do nothing (CR 608.2b), the tempt included.
// The creature may be chosen as the Ring-bearer: choosing it is not
// targeting.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60092fb4-98db-4f2d-a5d7-24b5f11c8acc",
		Name:         "Bombadil's Song",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Bombadil's Song — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"hexproof"}, Label: "Bombadil's Song — hexproof"}).Apply(ctx); err != nil {
					return err
				}
			}
			return TheRingTemptsYou{}.Apply(ctx)
		},
	})
}
