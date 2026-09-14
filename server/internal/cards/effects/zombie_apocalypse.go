package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zombie Apocalypse — Sorcery {3}{B}{B}{B} (EDHREC rank 3093):
//
//	"Return all Zombie creature cards from your graveyard to the
//	 battlefield tapped, then destroy all Humans."
//
// The Zombie deck's mass reanimation with a Human wipe on the side.
// Every Zombie creature card in the caster's graveyard comes back
// under its owner's control — the caster's, "your graveyard" — and
// is tapped; then every permanent with the Human subtype, anyone's,
// is destroyed as one simultaneous event (b29ReturnZombieCardsTappedThenDestroyHumans).
// A Zombie Human returns and then dies, as printed.
//
// One sandbox simplification, declared, weaker than printed: the
// Zombies enter untapped and are tapped a beat later (Splendid
// Reclamation's posture — ReturnFromGraveyard has no tapped flag), so
// anything watching for a tap event sees one. Both steps happen
// inside one resolution, so nothing gets a window in between.
//
// The second one this file used to declare — the mass-destroy path
// bypassing indestructible (#446), so an indestructible Human died
// where printed it would survive — was an engine gap, closed for
// every "destroy all" card at once in S30 (#470).
func init() {
	Register(Spec{
		OracleID:     "8241277d-654f-4985-9d49-a22c1e59eec2",
		Name:         "Zombie Apocalypse",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The Zombies enter untapped and are tapped immediately afterwards, so anything watching for a creature being tapped sees one.",
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b29ReturnZombieCardsTappedThenDestroyHumans(ctx)
		},
	})
}
