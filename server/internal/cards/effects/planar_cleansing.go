package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Planar Cleansing — Sorcery {3}{W}{W}{W}:
//
//	"Destroy all nonland permanents."
//
// The reset button. Six mana and triple white for the widest sweep
// white gets: creatures, artifacts, enchantments, planeswalkers and
// battles all go, and only the lands stay. A Commander table that
// has gone long enough for this to be castable usually has more on
// the battlefield than in hand, which is what makes it a blowout
// rather than a Wrath of God with extra steps.
//
// "Nonland permanent" is the whole predicate — Nonland() over the
// battlefield, where everything is by definition a permanent. It
// takes the caster's own board too; the card offers no exception and
// neither does this.
func init() {
	Register(Spec{
		OracleID: "a98c2d81-4add-4292-bbdd-e1b69ff936d4",
		Name:     "Planar Cleansing",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Nonland()}.Apply(ctx)
		},
	})
}
