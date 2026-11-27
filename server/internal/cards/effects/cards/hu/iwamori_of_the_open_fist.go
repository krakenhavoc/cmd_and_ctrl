package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Iwamori of the Open Fist — Legendary Creature — Human Monk {2}{G}{G},
// 5/5 (Betrayers of Kamigawa, S82 OLD, #5602):
//
//	"Trample
//	 When Iwamori of the Open Fist enters, each opponent may put a
//	 legendary creature card from their hand onto the battlefield."
//
// Wumpus Aberration's put-from-hand for every opponent in turn, each
// choosing for themselves; the creature enters under that opponent's
// control. An opponent with no legendary creature card in hand is never
// asked. Iwamori's own controller is not an opponent.
//
// No simplification.
func init() {
	const name = "Iwamori of the Open Fist"
	effects.Register(effects.Spec{
		OracleID:        "35734003-0c94-4a0c-9441-ab989e73bb11",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — each opponent may put a legendary creature card from their hand onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var opponents []uuid.UUID
					effects.EachOpponent(g, item.Controller, func(opp uuid.UUID) bool {
						opponents = append(opponents, opp)
						return false
					})
					for _, opp := range opponents {
						if err := (effects.PutFromHandOntoBattlefield{
							Player:   opp,
							Match:    effects.And(effects.Creature(), effects.Legendary()),
							Optional: true,
							Label:    name + " — you may put a legendary creature card from your hand onto the battlefield",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
