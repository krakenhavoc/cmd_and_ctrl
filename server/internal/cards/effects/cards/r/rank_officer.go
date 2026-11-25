package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rank Officer — Creature — Zombie Soldier {3}{B}, 3/1 (Modern Horizons):
//
//	"When this creature enters, you may discard a card. If you do, create a
//	 2/2 black Zombie creature token.
//	 {1}{B}, {T}, Exile a creature card from your graveyard: Each opponent loses
//	 2 life."
//
// The discard is the controller's own choice, asked as the trigger resolves
// ("you may" is the up-to-one floor of zero); the token comes only when a card
// was really discarded, through a keyed discard-then so the prompt is a restore
// point. The activated ability exiles the creature card as a cost (CR 118.3),
// so it can't be paid with an empty graveyard, and drains every opponent.
//
// No simplification.
const rankOfficerLabel = "Rank Officer — you may discard a card; if you do, create a 2/2 black Zombie creature token"

var rankOfficerDiscarded = game.RegisterDiscardThen("choose-cards/s82-old-b166-rank-officer-zombie", func(g *game.Game, r game.DiscardsDone) error {
	if !r.Discarded.Discarded(r.Player) {
		return nil
	}
	item := r.Item
	if item == nil {
		return nil
	}
	return effects.CreateToken{Controller: r.Player, Template: effects.TokenCard("2/2 black Zombie"), N: 1}.Apply(effects.NewContext(g, item))
})

func init() {
	effects.Register(effects.Spec{
		OracleID:     "52c73c88-3b87-4184-b8e1-cd2b5688800d",
		Name:         "Rank Officer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(rankOfficerLabel, func(g *game.Game, item *game.StackItem) error {
				return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
					Player:   item.Controller,
					Source:   item.SourceCardID,
					N:        1,
					UpTo:     true,
					Question: rankOfficerLabel,
					Asked:    game.NewAskedItem(item),
				}, rankOfficerDiscarded, nil)
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, {T}, Exile a creature card from your graveyard: Each opponent loses 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost(), effects.ExileFromGraveyard(1, "a creature card", effects.MatchCreature)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.EachOpponentLosesLifeNow(g, item, 2)
			},
		}},
	})
}
