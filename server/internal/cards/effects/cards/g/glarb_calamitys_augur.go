package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glarb, Calamity's Augur — Legendary Creature — Frog Wizard Noble {B}{G}{U}, 2/4 (Bloomburrow):
//
//	"Deathtouch
//	 You may look at the top card of your library any time.
//	 You may play lands and cast spells with mana value 4 or greater from the
//	 top of your library.
//	 {T}: Surveil 2."
//
// Augur of Autumn's shape: the top card is visible to you (LibraryTopOwner), and two
// standing permissions over it, one for lands (played, CR 305.1) and one for
// spells (cast, paying their costs) whose mana value must be at least four, the
// floor being CastPermission.MinSpellManaValue (the seam this card needed,
// MaxSpellManaValue's twin, judged in the same gate the cast path, the view's
// offers and the bot share, with X as zero on a library card). A land has no
// mana value to test and is never cast, so it stays on its own permission.
//
// No simplification.
func init() {
	four := 4
	spells := effects.PlayFromTopOfYourLibrary(
		game.PermissionFilter{NonLandOnly: true},
		"Cast a spell with mana value 4 or greater from the top of your library (Glarb, Calamity's Augur)")
	spells.MinSpellManaValue = &four
	effects.Register(effects.Spec{
		OracleID:          "df9409e6-bd6d-4b5b-b807-fcf347ee940a",
		Name:              "Glarb, Calamity's Augur",
		Completeness:      effects.CompletenessFull,
		PrintedKeywords:   []string{"deathtouch"},
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			effects.PlayFromTopOfYourLibrary(
				game.PermissionFilter{LandsOnly: true},
				"Play a land from the top of your library (Glarb, Calamity's Augur)"),
			spells,
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Surveil 2.",
			Cost:    effects.TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Surveil{Player: item.Controller, N: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
