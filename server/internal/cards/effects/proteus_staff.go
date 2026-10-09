package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Proteus Staff — Artifact {3}:
//
//	"{2}{U}, {T}: Put target creature on the bottom of its owner's
//	 library. That creature's controller reveals cards from the top of
//	 their library until they reveal a creature card. The player puts
//	 that card onto the battlefield and the rest on the bottom of their
//	 library in any order. Activate only as a sorcery."
//
// "That creature's controller" is read before the creature moves, as
// the ability resolves. The reveal runs whether or not the creature
// actually reached the library (a commander that took the command zone
// still left the battlefield). A library with no creature card reveals
// itself entirely and every revealed card goes to the bottom, which is
// the printed outcome. The creature enters under that player's
// control, since they put it there, and they order the rest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97024e8e-dfde-4769-bc27-c3d6fad19e7c",
		Name:         "Proteus Staff",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{2}{U}, {T}: Put target creature on the bottom of its owner's library. That creature's controller reveals cards from the top of their library until they reveal a creature card. The player puts that card onto the battlefield and the rest on the bottom of their library in any order. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}{U}"), TapCost()),
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature"),
			Effect:       proteusStaffResolve,
		}},
	})
}

func proteusStaffResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	player, ok := controllerOfTarget(ctx, id)
	if !ok {
		return nil
	}
	if err := g.TuckToLibraryForEffect(id, true); err != nil {
		return err
	}
	run, hit := revealUntil(ctx, player, func(c game.Card) bool { return c.IsCreature() },
		"Proteus Staff — reveal until a creature card")
	return PutFromLibraryOntoBattlefield{
		Player: player,
		Cards:  run,
		Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
		All:    true,
		Then: func(g *game.Game, res PutFromLibraryResult) error {
			return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
				Chooser:   player,
				Source:    res.Source,
				Cards:     res.Rest,
				From:      game.ZoneLibrary,
				Placement: game.LibraryPlaceBottom,
				Reason:    libraryOrderLabel(g, res.Source, "put the rest on the bottom of your library in any order"),
			})
		},
	}.Apply(ctx)
}
