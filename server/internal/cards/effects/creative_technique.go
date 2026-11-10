package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Creative Technique — Sorcery {4}{R}:
//
//	"Demonstrate (When you cast this spell, you may copy it. If you do,
//	 choose an opponent to also copy it.)
//	 Shuffle your library, then reveal cards from the top of it until
//	 you reveal a nonland card. Exile that card and put the rest on the
//	 bottom of your library in a random order. You may cast the exiled
//	 card without paying its mana cost."
//
// Demonstrate is the keyword's cast trigger (Demonstrate, CR 702.144a).
// Each copy resolves for its own controller: "your library" is the
// library of whoever controls the copy, so the opponent who copied it
// walks and casts from their own.
//
// The walk REVEALS (revealUntil): the cards passed over stay in the
// library and go to the bottom in a random order, and only the nonland
// card is exiled. The free cast is offered over it during the resolution
// (the resolution_cast prompt under FreeCastFromExilePermission); a card
// not cast stays in exile, as printed. A library with no nonland card is
// revealed whole and goes to the bottom, and nothing is cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c336b735-3e6b-471e-be7a-84d7402ae2f5",
		Name:         "Creative Technique",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Demonstrate()},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			you := ctx.Controller()
			if err := (ShuffleLibrary{Player: you}).Apply(ctx); err != nil {
				return err
			}
			run, hit := revealUntil(ctx, you, func(c game.Card) bool { return !c.IsLand() },
				"Creative Technique — reveal until a nonland card")
			var rest []uuid.UUID
			for _, id := range run {
				if id != hit {
					rest = append(rest, id)
				}
			}
			if err := ctx.Game.PutOnBottomInRandomOrderForEffect(you, game.ZoneLibrary, rest); err != nil {
				return err
			}
			if hit == uuid.Nil {
				return nil
			}
			const name = "Creative Technique"
			return exileWithPermissionThenOffer(ctx.Game, hit, item.SourceCardID,
				ctx.Game.FreeCastFromExilePermission(you, item.SourceCardID, name), name)
		},
	})
}
