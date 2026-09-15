package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wilderness Reclamation — Enchantment, {3}{G} (EDHREC rank 937):
//
//	"At the beginning of your end step, untap all lands you control."
//
// The instant-speed deck's second main phase: tap out on your turn,
// untap at the end step, hold up everything on everyone else's.
// One end-step trigger (yours only — the Actor check) over the same
// untap-a-set loop Sword of Feast and Famine uses. The mana pool
// empties between steps as usual, so the untapped lands are for the
// end step and beyond, not for the turn that just happened.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f856f99-4cb4-479d-958d-964220965ed6",
		Name:         "Wilderness Reclamation",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Wilderness Reclamation — untap all lands you control", func(g *game.Game, item *game.StackItem) error {
				return untapAllLandsControlledBy(g, item.Controller, NewContext(g, item))
			}),
		},
	})
}
