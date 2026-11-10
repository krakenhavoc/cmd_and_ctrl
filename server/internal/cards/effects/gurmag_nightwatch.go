package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gurmag Nightwatch — Creature — Human Ranger {2/B}{2/G}{2/U}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, look at the top three cards of your
//	 library. You may put one of those cards back on top of your library.
//	 Put the rest into your graveyard."
//
// A private look at the top three (a shorter library offers what it has),
// then one choose-cards prompt over them with a floor of zero: the card
// picked stays where it is and, with the rest gone, is on top; every other
// looked-at card still in the library is put (not milled) into the
// graveyard, so declining sends all of them. A card that left the library
// while the prompt was open is a new object and stays where it went
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fde50a0d-9bc6-45f4-873e-3de81a513fac",
		Name:         "Gurmag Nightwatch",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gurmag Nightwatch — look at the top three cards of your library; you may put one back on top and the rest into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					looked := g.LookAtTopOfLibraryForEffect(item.Controller, 3)
					if len(looked) == 0 {
						return nil
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:  item.Controller,
						Source:   item.SourceCardID,
						Question: "Gurmag Nightwatch — you may put one of those cards back on top of your library; the rest go into your graveyard",
						Cards:    looked,
						Min:      0,
						Max:      1,
						LiveZone: game.ZoneLibrary,
						ThenKey:  gurmagNightwatchPick,
						Asked:    game.NewAskedItem(item),
					})
					return nil
				}),
		},
	})
}

// gurmagNightwatchPick puts every looked-at card that is still in the
// library and was not chosen into the graveyard. An on-disk key: never
// renamed, never reused.
var gurmagNightwatchPick = askedPick("choose-cards/gurmag-nightwatch-rest", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Candidates {
		if slices.Contains(r.Picked, id) || !stdB9InLibrary(ctx.Game, id) {
			continue
		}
		if err := ctx.Game.PutIntoGraveyardForEffect(id); err != nil {
			return err
		}
	}
	return nil
})
