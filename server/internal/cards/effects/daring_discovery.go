package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daring Discovery — Sorcery {4}{R}:
//
//	"Up to three target creatures can't block this turn.
//	 Discover 4."
//
// The block restriction is the until-end-of-turn "can't" record on each
// target still legal as the spell resolves (CR 608.2b); the discover
// happens whether or not any target was left, because it is a separate
// sentence. Discover is ADR 0099's (game/discover.go).
func init() {
	Register(Spec{
		OracleID:     "a9564593-b5a6-4a83-be1e-2af8caf647d7",
		Name:         "Daring Discovery",
		Completeness: CompletenessFull,
		Discovers:    true,
		Targets:      TargetCreature("up to three target creatures").WithCount(0, 3),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: "Daring Discovery — can't block"}).Apply(ctx); err != nil {
					return err
				}
			}
			return Discover{N: 4}.Apply(ctx)
		},
	})
}
