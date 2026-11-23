package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tromp the Domains — Sorcery {5}{G} (Time Spiral, #5596):
//
//	"Domain — Until end of turn, creatures you control gain trample and
//	 get +1/+1 for each basic land type among lands you control."
//
// The domain count (DomainCount) is read once as the spell resolves, and
// the creatures you control then are the ones affected (CR 611.2c): each
// gains trample and gets +N/+N until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b5dece6-b1e0-4844-bbf9-3b44008ffa70",
		Name:         "Tromp the Domains",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			yours := effects.And(effects.Creature(), effects.YouControl())
			if err := (effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"trample"},
				Label: "Tromp the Domains — trample until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			n := effects.DomainCount(ctx.Game, ctx.Controller())
			return effects.BoostUntilEOT{Match: yours, Power: n, Toughness: n,
				Label: "Tromp the Domains — +1/+1 for each basic land type"}.Apply(ctx)
		},
	})
}
