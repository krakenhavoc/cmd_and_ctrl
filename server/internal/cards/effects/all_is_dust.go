package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// All Is Dust — Kindred Sorcery — Eldrazi, {7} (EDHREC rank 1001):
//
//	"Each player sacrifices all permanents they control that are one
//	 or more colors."
//
// The colorless deck's one-sided wipe: everything with a colour goes,
// and nothing of yours does. A SACRIFICE, not a destruction, which is
// the whole reason to play it over a Wrath — indestructible does not
// save anything, and every death is a sacrifice for the cards that
// watch for one. "One or more colors" reads the same colour surface
// every colour predicate does (stamped colours, mana-cost fallback),
// so lands and colorless artifacts stay and a red Goblin token goes.
//
// The set is snapshotted before anything moves (CR 608.2) and then
// sacrificed as ONE simultaneous exit, so a coloured Blood Artist
// swept by the spell sees every death including its own (CR 603.10).
// That was a declared simplification until #910 gave sacrifice the
// batch the destroy, exile and bounce sweeps already had.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "14693689-d087-43b6-9c3f-63ab0648fc20",
		Name:         "All Is Dust",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b08SacrificeAllMatching(ctx, Not(Colorless()))
		},
	})
}
