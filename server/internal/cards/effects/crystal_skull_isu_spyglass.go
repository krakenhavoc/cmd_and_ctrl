package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crystal Skull, Isu Spyglass — Legendary Artifact {2}{U}{U}:
//
//	"You may look at the top card of your library any time.
//	 You may play historic lands and cast historic spells from the top
//	 of your library. (Artifacts, legendaries, and Sagas are historic.)
//	 {T}: Add {U}."
//
// The Bolas's Citadel / Realmwalker shape: the look is the PRIVATE
// strength (LibraryTopOwner — "you may look", never "revealed"), and
// the permission is PlayFromTopOfYourLibrary narrowed by
// PermissionFilter.HistoricOnly. The filter deliberately does not set
// LandsOnly or a spell-only flag: the printed clause is "historic
// lands and historic spells", which is every historic card, and a land
// is played (CR 305.1) while a spell is cast, both through the same
// library permission. A non-historic card on top is visible to you and
// stays unplayable. Historic is artifact, legendary or Saga (CR 700.6),
// read from the card's effective characteristics. The permission
// charges the printed cost, there is no per-turn limit, and the
// permission is derived from the battlefield, so it is gone the moment
// the Skull leaves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "b1f8eb7b-7e01-4bf8-a557-64e16a6052af",
		Name:              "Crystal Skull, Isu Spyglass",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{HistoricOnly: true},
				"Play a historic land or cast a historic spell from the top of your library (Crystal Skull, Isu Spyglass)"),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
	})
}
