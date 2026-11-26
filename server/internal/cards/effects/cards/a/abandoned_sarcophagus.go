package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abandoned Sarcophagus — {3} Artifact (Commander 2020):
//
//	"You may cast spells that have a cycling ability from your graveyard.
//	 If a card that has a cycling ability would be put into your graveyard from
//	 anywhere and it wasn't cycled, exile it instead."
//
// The first line is a standing graveyard cast permission whose filter reads
// the card's cycling ability (PermissionFilter.HasCyclingOnly); CastOnly keeps
// a cycling land (a Triome) from being offered, as it is not a spell. The
// second is a replacement in the CR 614 window of every move to the
// controller's graveyard (S82OldB122ExileCyclingCardsNotCycled). A card cycled
// goes to the graveyard, where the first line lets it be cast.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a051dee0-60c8-4f58-84cb-55460c097115",
		Name:         "Abandoned Sarcophagus",
		Completeness: effects.CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:     game.ZoneGraveyard,
			Filter:   game.PermissionFilter{HasCyclingOnly: true},
			CastOnly: true,
			Label:    "Cast a spell that has a cycling ability from your graveyard (Abandoned Sarcophagus)",
		}},
		Replacements: []game.ReplacementEffect{
			effects.S82OldB122ExileCyclingCardsNotCycled("Abandoned Sarcophagus — a card with cycling that wasn't cycled is exiled instead of going to your graveyard"),
		},
	})
}
