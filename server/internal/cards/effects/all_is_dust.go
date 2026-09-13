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
// Sandbox simplification: the set is snapshotted before anything
// moves (CR 608.2), but the permanents are then sacrificed one at a
// time rather than as one simultaneous event — the engine's
// simultaneous-exit batch covers destroy, exile and bounce, not
// sacrifice. A coloured dies-watcher swept by the spell (a Blood
// Artist) therefore sees only the permanents that leave after it
// does, which is WEAKER than printed for its controller and never
// stronger.
func init() {
	Register(Spec{
		OracleID:     "14693689-d087-43b6-9c3f-63ab0648fc20",
		Name:         "All Is Dust",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Permanents are sacrificed one after another rather than all at once, so a creature swept by the spell doesn't see the ones sacrificed after it."},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b08SacrificeAllMatching(ctx, Not(Colorless()))
		},
	})
}
