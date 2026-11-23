package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thought Sponge — Creature — Sponge {3}{U}, 1/1 (Commander 2019, #5517):
//
//	"Flash
//	 This creature enters with a number of +1/+1 counters on it equal to
//	 the greatest number of cards an opponent has drawn this turn.
//	 When this creature dies, draw cards equal to its power."
//
// The entry counters are Centaur Vinecrasher's counted "enters with" clause
// (CR 614.1c), read as the Sponge enters: the largest of the turn's draw
// tallies among its controller's opponents (Game.DrawnThisTurn, which counts
// every card drawn this turn whether or not it is still in hand). The death
// trigger is Doom Weaver's: cards equal to the Sponge's last-known power
// (CR 608.2h), nothing at zero or less.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "256e094a-a9e8-42d9-90e2-7a24ed85e426",
		Name:            "Thought Sponge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, c19MostCardsAnOpponentDrewThisTurn,
				"Thought Sponge: enters with a +1/+1 counter for each card the opponent who drew the most cards this turn drew"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Thought Sponge — draw cards equal to its power", effects.DoomWeaverDraw),
		},
	})
}

// c19MostCardsAnOpponentDrewThisTurn is the greatest number of cards any one
// opponent of src's controller has drawn this turn.
func c19MostCardsAnOpponentDrewThisTurn(g *game.Game, src *game.Card) int {
	most := 0
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == src.Controller {
			continue
		}
		if n := len(g.DrawnThisTurn[p.ID]); n > most {
			most = n
		}
	}
	return most
}
