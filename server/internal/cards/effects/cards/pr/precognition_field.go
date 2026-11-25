package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Precognition Field — Enchantment {3}{U} (Dominaria, S82 OLD, #5530):
//
//	"You may look at the top card of your library any time.
//	 You may cast instant and sorcery spells from the top of your library.
//	 {3}: Exile the top card of your library."
//
// Mystic Forge's shape: the private look (LibraryTopOwner, CR 401.5), a
// cast permission over the top card filtered to instants and sorceries
// (costs and timing are the card's own), and an exile that takes whatever is
// on top as the ability resolves. The exile is public and is not a draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "9c3e4bd4-2e1e-4103-adc3-f9ccc259dcdf",
		Name:              "Precognition Field",
		Completeness:      effects.CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			effects.PlayFromTopOfYourLibrary(game.PermissionFilter{InstantOrSorceryOnly: true},
				"Cast an instant or sorcery spell from the top of your library (Precognition Field)"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Exile the top card of your library.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.MillToZone{Player: ctx.Controller(), N: 1, To: game.ZoneExile}.Apply(ctx)
			},
		}},
	})
}
