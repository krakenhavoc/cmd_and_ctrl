package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Maze's End — Land (Foundations):
//
//	"This land enters tapped.
//	 {T}: Add {C}.
//	 {3}, {T}, Return this land to its owner's hand: Search your library
//	 for a Gate card, put it onto the battlefield, then shuffle. If you
//	 control ten or more Gates with different names, you win the game."
//
// The return is a cost (ReturnThis), so the search resolves with Maze's
// End already in hand. The win is the search's keyed tail: it runs once
// the Gate is on the battlefield (and after the shuffle), counting the
// distinct names among the Gates its controller controls then (CR 608.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49479778-c4c0-43ba-a7b7-45f00d067462",
		Name:         "Maze's End",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Return this land to its owner's hand: Search your library for a Gate card, put it onto the battlefield, then shuffle. If you control ten or more Gates with different names, you win the game.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.ReturnThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.HasSubtypeWord("Gate") },
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Reason:    "Maze's End — a Gate card",
					Source:    item.SourceCardID,
					ThenTail:  ctx.ItemTail(mazesEndWin),
				}.Apply(ctx)
			},
		}},
	})
}

// mazesEndWin is "if you control ten or more Gates with different names,
// you win the game". An on-disk key.
var mazesEndWin = effects.CardTail("mazes-end/win-the-game", func(ctx *effects.Context, o game.TailOutcome) error {
	names := map[string]bool{}
	for _, id := range effects.PermanentsControlledByMatching(ctx.Game, o.Tail.Player, effects.HasSubtype("Gate")) {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			names[c.Name] = true
		}
	}
	if len(names) < 10 {
		return nil
	}
	return effects.WinTheGame{Player: o.Tail.Player}.Apply(ctx)
})
