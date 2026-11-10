package effects

import (
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
	Register(Spec{
		OracleID:     "49479778-c4c0-43ba-a7b7-45f00d067462",
		Name:         "Maze's End",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}, Return this land to its owner's hand: Search your library for a Gate card, put it onto the battlefield, then shuffle. If you control ten or more Gates with different names, you win the game.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}"), TapCost(), ReturnThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return SearchLibrary{
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
var mazesEndWin = CardTail("mazes-end/win-the-game", func(ctx *Context, o game.TailOutcome) error {
	names := map[string]bool{}
	for _, id := range permanentsControlledByMatching(ctx.Game, o.Tail.Player, HasSubtype("Gate")) {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			names[c.Name] = true
		}
	}
	if len(names) < 10 {
		return nil
	}
	return WinTheGame{Player: o.Tail.Player}.Apply(ctx)
})
