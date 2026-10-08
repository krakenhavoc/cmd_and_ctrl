package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Lady of Otaria — {3}{R}{G} Legendary Creature — Avatar 5/5
// (#2664):
//
//	"You may tap three untapped Dwarves you control rather than pay
//	 this spell's mana cost.
//	 At the beginning of each end step, if a land you controlled was put
//	 into a graveyard from the battlefield this turn, reveal the top four
//	 cards of your library. You may put any number of Dwarf cards from
//	 among them into your hand. Put the rest on the bottom of your
//	 library in a random order."
//
// The alternative cost is ADR 0135 §1's TapInstead with no condition.
//
// The trigger fires at EACH end step, so a four-player table gives it
// four chances a turn cycle. "A land you controlled" is the Lady's
// controller, read off the per-player tally cell
// game.PlayerTurnTally.LandsToGraveyard (#2664): a land that went from
// the battlefield to a graveyard under that player this turn, by any
// route. It is an intervening "if" (CR 603.4), so it is checked as the
// end step begins and again as the trigger resolves. The cell resets
// with the rest of the turn tally.
//
// The four cards are revealed to the table, the Dwarves taken go to
// hand, and the rest go to the bottom in a random order, as Artificer
// Class does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ea4f882-6872-4d9e-9532-5ba2bf848d00",
		Name:         "The Lady of Otaria",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			TapInstead(3, "three untapped Dwarves you control", nil, HasSubtype("Dwarf")),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return g.LandToGraveyardThisTurn(source.Controller)
			}, "The Lady of Otaria — reveal the top four cards of your library",
				func(g *game.Game, item *game.StackItem) error {
					// CR 603.4: re-checked on resolution.
					if !g.LandToGraveyardThisTurn(item.Controller) {
						return nil
					}
					ctx := NewContext(g, item)
					var revealed []uuid.UUID
					if err := (RevealTopOfLibrary{
						Player:   item.Controller,
						N:        4,
						Reason:   "The Lady of Otaria — reveal the top four cards of your library",
						Revealed: &revealed,
					}).Apply(ctx); err != nil {
						return err
					}
					return TakeFromLibraryToHand{
						Player:   item.Controller,
						Cards:    revealed,
						Match:    HasSubtype("Dwarf"),
						Optional: true,
						Label:    "The Lady of Otaria — you may put any number of Dwarf cards into your hand",
						Then:     TakeRestOnBottomInRandomOrder,
					}.Apply(ctx)
				}),
		},
	})
}
