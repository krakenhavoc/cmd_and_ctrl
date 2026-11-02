package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Madame Web, Clairvoyant — Legendary Creature — Mutant Advisor {4}{U}{U},
// 4/4 (Marvel's Spider-Man, #4460):
//
//	"You may look at the top card of your library any time.
//	 You may cast Spider spells and noncreature spells from the top of your
//	 library.
//	 Whenever you attack, you may mill a card."
//
// The look is the private top-card designation (Daredevil's
// LibraryTopOwner). The permission is the Vizier of the Menagerie / Elven
// Chorus play-from-the-top grant with Scarlet Witch's filter: "a Spider or
// noncreature spell" is PermissionFilter.NoncreatureOrCreatureType "Spider",
// judged on the spell the cast would make, so an instant or an Aura passes,
// a Spider creature passes and any other creature is refused. NonLandOnly
// and CastOnly keep a land on top from being played: a land is never cast,
// and "Spider spells and noncreature spells" names spells only. The printed
// mana cost is paid. "Whenever you attack" is ONE trigger per attack
// declaration, however many creatures attack (OncePerBatch), and the "you
// may" is asked as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "b1e0841d-b004-40f9-a644-aecb422f6ad5",
		Name:              "Madame Web, Clairvoyant",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions:   []game.CastPermission{madameWebCastFromTop()},
		Triggered: []game.TriggeredAbility{
			Optional(OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Madame Web, Clairvoyant — mill a card", Do(MillCards{N: 1}))),
				"Madame Web, Clairvoyant — mill a card?"),
		},
	})
}

// madameWebCastFromTop is "you may cast Spider spells and noncreature spells
// from the top of your library".
func madameWebCastFromTop() game.CastPermission {
	p := PlayFromTopOfYourLibrary(
		game.PermissionFilter{NoncreatureOrCreatureType: "Spider", NonLandOnly: true},
		"Cast a Spider spell or a noncreature spell from the top of your library (Madame Web, Clairvoyant)")
	p.CastOnly = true
	return p
}
