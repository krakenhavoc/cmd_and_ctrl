package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Fourth Doctor — Legendary Creature — Time Lord Doctor {2}{G}{U}, 4/4
// (Doctor Who):
//
//	"You may look at the top card of your library any time.
//	 Would You Like A...? — Once each turn, you may play a historic land or cast a
//	 historic spell from the top of your library. When you do, create a Food token.
//	 (Artifacts, legendaries, and Sagas are historic.)"
//
// Johann's shape over Crystal Skull's filter: the look is the private strength
// (LibraryTopOwner) and the standing permission is a historic play from the top of
// your library with one use per turn, recorded per granting object on the turn
// tally (PerType, so a Doctor that leaves and a new one each give a fresh use).
// The Food trigger is "when you do": a historic spell you cast from the top of your
// library (the cast leaves the library), or a historic land you PLAY from it
// (Event.Played, CR 305.4 — a land put onto the battlefield by an effect is not a
// play). It triggers however the card was allowed to be played from the top, as the
// printed words read.
//
// No simplification.
func init() {
	const label = "Would You Like A...? — create a Food token"
	perm := effects.PlayFromTopOfYourLibrary(game.PermissionFilter{HistoricOnly: true},
		"Once each turn, play a historic land or cast a historic spell from the top of your library (The Fourth Doctor)")
	perm.PerType = []string{game.PermissionTypePermanent}
	effects.Register(effects.Spec{
		OracleID:          "c19fdee3-5fbc-4e2d-ac23-85b46eacafdc",
		Name:              "The Fourth Doctor",
		Completeness:      effects.CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		Purpose:           game.Purpose{Tokens: 1},
		CastPermissions:   []game.CastPermission{perm},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventCast, game.EventZoneMove}, effects.GapB86HistoricPlayedFromYourLibrary,
				label, effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
	})
}
