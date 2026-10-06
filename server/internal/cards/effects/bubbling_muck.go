package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bubbling Muck — Sorcery {B}:
//
//	"Until end of turn, whenever a player taps a Swamp for mana, that
//	 player adds an additional {B}."
//
// A triggered mana ability (CR 605.1b) set up for the turn, so it is a
// TurnManaTrigger rather than a stack trigger: every player's Swamp, the
// caster's included, for the rest of the turn, resolving at once. It is
// not tied to the sorcery, which is in the graveyard by then, and it
// ends at cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c7f2b29-9f42-41ab-a2d4-7a450fb0242d",
		Name:         "Bubbling Muck",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return TurnManaTrigger{
				Label:   "Bubbling Muck — add an additional {B}",
				Subtype: "Swamp",
				Adds:    "{B}",
			}.Apply(ctx)
		},
	})
}
