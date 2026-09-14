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
// Two sandbox simplifications, declared, both weaker than printed:
//
//   - The Zombies enter untapped and are tapped a beat later
//     (Splendid Reclamation's posture — ReturnFromGraveyard has no
//     tapped flag), so anything watching for a tap event sees one.
//     Both steps happen inside one resolution, so nothing gets a
//     window in between.
//   - The mass-destroy path bypasses indestructible (#446), so an
//     indestructible Human is destroyed where printed it would
//     survive — the same caveat every "destroy all" card carries.
func init() {
	Register(Spec{
		OracleID:     "8241277d-654f-4985-9d49-a22c1e59eec2",
		Name:         "Zombie Apocalypse",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The Zombies enter untapped and are tapped immediately afterwards, so anything watching for a creature being tapped sees one.",
			"Indestructible saves a permanent from single-target removal and from lethal damage, but a board wipe (\"destroy all\") still destroys it.",
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b29ReturnZombieCardsTappedThenDestroyHumans(ctx)
		},
	})
}
