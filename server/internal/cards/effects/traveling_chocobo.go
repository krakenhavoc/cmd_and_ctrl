package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Traveling Chocobo — Creature — Bird {2}{G}, 3/2 (Final Fantasy,
// #4462):
//
//	"You may look at the top card of your library any time.
//	 You may play lands and cast Bird spells from the top of your
//	 library.
//	 If a land or Bird you control entering the battlefield causes a
//	 triggered ability of a permanent you control to trigger, that
//	 ability triggers an additional time."
//
// The look is the private strength (LibraryTopOwner), as on Vizier of the
// Menagerie; the permission is two top-of-library grants (Courser of
// Kruphix's land play, Gisa and Geralf's typed cast), each at the printed
// cost. The last line is Ancient Greenwarden's doubler (DoublesEntering)
// over a land or a Bird that is under YOUR control as it enters — narrower
// than Panharmonicon's, which asks nothing of the entering object's
// controller. The Chocobo is a Bird, so its own entry and the Birds that
// follow it count, and the doubler is for triggers of permanents you
// control.
//
// No simplification. The card declares LandfallExtraTriggers 1, the
// doubler as the heuristic reads it beside the landfall payoffs.
func init() {
	doubler := DoublesEntering(And(YouControl(), Or(Land(), OfCreatureType("Bird"))))
	doubler.Label = "Traveling Chocobo"
	doubler.RowLabel = "Abilities that trigger when a land or Bird you control enters trigger an additional time."
	Register(Spec{
		OracleID:          "2e5675d3-3fb0-43b5-a65c-95aacbe48bf9",
		Name:              "Traveling Chocobo",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		Purpose:           game.Purpose{LandfallExtraTriggers: 1},
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(game.PermissionFilter{LandsOnly: true},
				"Play a land from the top of your library (Traveling Chocobo)"),
			PlayFromTopOfYourLibrary(game.PermissionFilter{CreatureOnly: true, CreatureType: "Bird"},
				"Cast a Bird spell from the top of your library (Traveling Chocobo)"),
		},
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})
}
