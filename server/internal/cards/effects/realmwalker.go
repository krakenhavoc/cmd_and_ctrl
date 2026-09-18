package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Realmwalker — Creature — Shapeshifter {2}{G}, 2/3:
//
//	"Changeling
//	 As this creature enters, choose a creature type.
//	 You may look at the top card of your library any time.
//	 You may cast creature spells of the chosen type from the top of
//	 your library."
//
// The library permission that reads a per-instance CHOICE, and the
// reason PermissionFilter carries FromChosenType rather than a fixed
// subtype: two Realmwalkers name two different types, and the filter
// has to be derived from the permanent granting it rather than from
// the catalog entry. The derivation reads S26's Card.NamedTribe —
// which is nothing at all until the entry choice is answered (CR
// 614.12), and a Realmwalker with no type named grants no permission
// at all rather than one that opens everything.
//
// "Cast creature SPELLS", not "play cards", so the permission opens no
// land: a land on top under a Realmwalker is visible and unplayable,
// which is the printed card.
//
// The visibility is private (CR 401.5's "you may look at the top card
// of your library any time"), so only the Realmwalker's controller
// sees what is on top — the difference from Courser of Kruphix, and
// the one that decides whether the opponent can plan around it.
//
// Changeling (CR 702.73) rides PrintedKeywords and is what makes the
// card read oddly well against itself: a changeling in your library is
// every creature type, so a Realmwalker naming any type can cast a
// second Realmwalker off the top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "b81eaa2f-0554-41c6-bdf6-d1cb73b8f56f",
		Name:              "Realmwalker",
		Completeness:      CompletenessFull,
		PrintedKeywords:   []string{"changeling"},
		AsEnters:          ChooseCreatureTypeAsEnters("Realmwalker"),
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{CreatureOnly: true, FromChosenType: true},
				"Cast a creature spell of the chosen type from the top of your library (Realmwalker)"),
		},
	})
}
