package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zombie Apocalypse — Sorcery {3}{B}{B}{B} (EDHREC rank 3093):
//
//	"Return all Zombie creature cards from your graveyard to the
//	 battlefield tapped, then destroy all Humans."
//
// The Zombie deck's mass reanimation with a Human wipe on the side.
// Every Zombie creature card in the caster's graveyard comes back
// under its owner's control — the caster's, "your graveyard" — all
// of them in one entry and tapped as they enter (#1867); then every
// permanent with the Human subtype, anyone's, is destroyed as one
// simultaneous event (b29ReturnZombieCardsTappedThenDestroyHumans).
// A Zombie Human returns and then dies, as printed. The destroy waits
// for the return to finish, so a Zombie whose entry stops to ask a
// question is on the battlefield before the Humans die.
//
// No simplification. The one this file used to declare, the Zombies
// entering untapped and being tapped a beat later, went with #1867.
// The other — the mass-destroy path bypassing indestructible (#446),
// so an indestructible Human died where printed it would survive — was
// an engine gap, closed for every "destroy all" card at once in S30
// (#470).
func init() {
	Register(Spec{
		OracleID: "8241277d-654f-4985-9d49-a22c1e59eec2",
		Name:     "Zombie Apocalypse",
		// ADR 0126 §6: only Humans.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b29ReturnZombieCardsTappedThenDestroyHumans(ctx)
		},
	})
}
