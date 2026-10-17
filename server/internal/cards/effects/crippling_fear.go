package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crippling Fear — Sorcery {2}{B}{B}:
//
//	"Choose a creature type. Creatures that aren't of the chosen type
//	 get -3/-3 until end of turn."
//
// Kindred Dominance's shape with a shrink in place of a destroy (#4092,
// Captain N'ghathrod's list). The type is chosen as the spell resolves
// (#2382, CR 700.2a), by the caster. Every creature that is not of that
// type shrinks, yours included; a changeling is of every type and is
// never shrunk, whatever is named (notOfCreatureType reads effective
// subtypes, CR 702.73a).
//
// The set locks when the effect begins (CR 611.2c): a creature that
// enters later this turn is not shrunk, and one that is shrunk stays
// shrunk even if it later becomes of the chosen type. The -3/-3 is a
// layer 7c modification, so it stacks with anthems and counters, and a
// creature taken to 0 toughness dies to the state-based action.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e7a7ba65-ad14-41d5-889b-9bbeae9ab3f7",
		Name:         "Crippling Fear",
		Completeness: CompletenessFull,
		// Creatures of the chosen type stay out of it (Partial).
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 3, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ChooseCreatureTypeThen(ctx.Game, item.Controller, item.SourceCardID, "Crippling Fear — choose a creature type",
				func(g *game.Game, t string) error {
					if t == "" {
						return nil
					}
					return BoostUntilEOT{Match: notOfCreatureType(t), Power: -3, Toughness: -3, Label: "Crippling Fear — -3/-3"}.Apply(NewContext(g, item))
				})
			return nil
		},
	})
}
