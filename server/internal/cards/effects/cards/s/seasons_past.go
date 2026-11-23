package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seasons Past — Sorcery {4}{G}{G} (Tales of Middle-earth Commander):
//
//	"Return any number of cards with different mana values from your
//	 graveyard to your hand. Put Seasons Past on the bottom of its owner's
//	 library."
//
// The cards are chosen from your graveyard as the spell resolves (it does not
// target), as a set whose mana values are all different (a registered set
// rule over the picked cards, read in the graveyard where X is zero; a split
// card counts by its combined mana value). Seasons Past is put on the bottom
// of its owner's library as part of its own resolution (not a replacement of
// where it goes), so a Seasons Past that is countered goes to the graveyard
// as any spell does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c46738a4-e0d4-41dc-a20d-4800f00cea62",
		Name:         "Seasons Past",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			cands := effects.GraveyardCardIDs(ctx, ctx.Controller(), func(game.Card) bool { return true })
			if len(cands) == 0 {
				return s82GapB120SeasonsPastTuck(ctx)
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:    ctx.Controller(),
				FromPlayer: ctx.Controller(),
				Source:     item.SourceCardID,
				Question:   "Seasons Past — return any number of cards with different mana values from your graveyard to your hand",
				Cards:      cands,
				Min:        0,
				Max:        len(cands),
				LiveZone:   game.ZoneGraveyard,
				ThenKey:    s82GapB120SeasonsPastPicked,
				Asked:      game.NewAskedItem(item),
			})
			return nil
		},
	})
}

// s82GapB120DifferentManaValues is the set rule: no two picked cards share a
// mana value.
func s82GapB120DifferentManaValues(picked []game.Card) bool {
	seen := map[int]bool{}
	for _, c := range picked {
		mv := c.ManaValue()
		if seen[mv] {
			return false
		}
		seen[mv] = true
	}
	return true
}

var s82GapB120SeasonsPastPicked = effects.AskedPickRule("choose-cards/seasons-past-different-mana-values", s82GapB120DifferentManaValues,
	func(ctx *effects.Context, r game.CardsChosen) error {
		for _, id := range r.Picked {
			if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
		}
		return s82GapB120SeasonsPastTuck(ctx)
	})

// s82GapB120SeasonsPastTuck puts the resolving Seasons Past on the bottom of
// its owner's library, if it is still the spell on the stack.
func s82GapB120SeasonsPastTuck(ctx *effects.Context) error {
	id := ctx.Item.SourceCardID
	if id == uuid.Nil {
		return nil
	}
	return ctx.Game.TuckToLibraryForEffect(id, true)
}
