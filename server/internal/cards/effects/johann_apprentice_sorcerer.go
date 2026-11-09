package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Johann, Apprentice Sorcerer — Legendary Creature — Human Wizard Sorcerer
// {2}{U}{R}, 4/4 (Wilds of Eldraine, #4471):
//
//	"You may look at the top card of your library any time.
//	 Once each turn, you may cast an instant or sorcery spell from the top
//	 of your library. (You still pay its costs. Timing rules still apply.)"
//
// The first sentence is the visible-library-top declaration (Bolas's
// Citadel, Courser of Kruphix: LibraryTopOwner). The second is a standing
// cast permission over the top card with Serra Paragon's one use per turn,
// PerType PermissionTypePermanent, which asks no question and is recorded per
// granting object on the turn tally (CR 400.7): a Johann that leaves and a
// new one each grant a fresh use that turn (the ruling), and a Johann that
// changes hands opens its new controller's own. The filter judges the spell,
// so an adventurer card on top may be cast as its instant or sorcery
// Adventure (the ruling) and not as its creature. Costs and timing are the
// card's own: nothing is waived.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "8068f80f-88ee-4c0b-a0fb-3a6337f83f22",
		Name:              "Johann, Apprentice Sorcerer",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{func() game.CastPermission {
			p := PlayFromTopOfYourLibrary(game.PermissionFilter{InstantOrSorceryOnly: true},
				"Once each turn, cast an instant or sorcery spell from the top of your library (Johann, Apprentice Sorcerer)")
			p.PerType = []string{game.PermissionTypePermanent}
			p.CastOnly = true
			return p
		}()},
	})
}
