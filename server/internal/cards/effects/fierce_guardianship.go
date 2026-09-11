package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fierce Guardianship — Instant {2}{U}:
//
//	"If you control a commander, you may cast this spell without
//	 paying its mana cost.
//	 Counter target noncreature spell."
//
// The most-played of the Commander Legends free spells: a Negate that
// costs nothing while your commander is out, which is the difference
// between holding up interaction and tapping out.
//
// S28: the alternative cost is now real. It shipped in S22 as a plain
// {2}{U} Negate — strictly weaker than printed — because the engine's
// CR 118.9 machinery had no way to express a CONDITIONAL offer. It
// does now: AlternativeCost.Condition gates the offer in the view as
// well as at announce, so a player with no commander on the
// battlefield is never shown a button the server would reject.
//
// "You control a commander" is a permanent you CONTROL, not one you
// own: a commander sitting in the command zone does not switch this
// on, and a commander you have stolen does.
//
// Deadly Rollick and the rest of the cycle can follow the same
// two-line pattern — see docs/decklists/top-100-commander-staples.md
// for the list.
func init() {
	Register(Spec{
		OracleID: "d09c9cba-fdd2-479b-ad5d-d05181c3e3f9",
		Name:     "Fierce Guardianship",
		Targets:  TargetSpell("target noncreature spell", Noncreature()),
		AlternativeCosts: []game.AlternativeCost{
			FreeIfYouControlCommander("Cast without paying its mana cost (you control a commander)"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
