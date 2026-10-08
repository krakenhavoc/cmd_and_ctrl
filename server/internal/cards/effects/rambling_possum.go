package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rambling Possum — Creature — Possum Mount {2}{G}:
//
//	"Whenever this creature attacks while saddled, it gets +1/+2 until
//	 end of turn. Then you may return any number of creatures that
//	 saddled it this turn to their owner's hand.
//	 Saddle 1"
//
// The only proof card that reads "creatures that saddled it this turn"
// (CR 702.171c): the candidates are Game.SaddlersOf, taken when the
// trigger resolves, so a saddler that has since left the battlefield is
// not offered, and the choice is a card-set prompt with a floor of zero
// ("any number" includes none). The pump comes first and does not wait
// for the prompt, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b100e6f-6c19-49d2-b9c4-613d63d42151",
		Name:         "Rambling Possum",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Saddle(1)},
		Triggered: []game.TriggeredAbility{
			AttacksWhileSaddled("Rambling Possum — +1/+2, then return the creatures that saddled it", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 2, Label: "Rambling Possum — +1/+2"}).Apply(ctx); err != nil {
					return err
				}
				saddlers := g.SaddlersOf(item.SourceCardID)
				if len(saddlers) == 0 {
					return nil
				}
				g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:  item.Controller,
					Source:   item.SourceCardID,
					Question: "Rambling Possum — return any number of the creatures that saddled it to their owners' hands",
					Cards:    saddlers,
					Min:      0,
					Max:      len(saddlers),
					Zone:     game.ZoneBattlefield,
					Then: func(g *game.Game, picked []uuid.UUID) error {
						for _, id := range picked {
							if err := (BounceToHand{Target: id}).Apply(NewContext(g, item)); err != nil {
								return err
							}
						}
						return nil
					},
				})
				return nil
			}),
		},
	})
}
