package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teferi's Puzzle Box — Artifact {4} (Dominaria Remastered):
//
//	"At the beginning of each player's draw step, that player puts the
//	 cards in their hand on the bottom of their library in any order,
//	 then draws that many cards."
//
// One trigger on every player's draw step. The player whose step it is
// does the putting (their own ordering prompt) and the drawing; the
// number drawn is the number of cards that actually reached the library,
// carried to the tail as data. With an empty hand nothing is put and
// nothing is drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37abcc92-9466-47ea-9e0b-5eda2eb62c8e",
		Name:         "Teferi's Puzzle Box",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginDrawStep, effects.AnyPlayer,
				"Teferi's Puzzle Box — that player puts the cards in their hand on the bottom of their library in any order, then draws that many cards",
				teferisPuzzleBoxResolve),
		},
	})
}

func teferisPuzzleBoxResolve(g *game.Game, item *game.StackItem) error {
	player := effects.TriggeringActor(item)
	if player == uuid.Nil {
		return nil
	}
	hand := effects.AllHandCardIDs(g, player)
	if len(hand) == 0 {
		return nil
	}
	ctx := effects.NewContext(g, item)
	then := ctx.Tail(teferisPuzzleBoxDraw)
	then.Player = player
	then.N = len(hand)
	return effects.PutInLibraryInAnyOrder{
		Chooser:   player,
		Cards:     hand,
		From:      game.ZoneHand,
		Placement: game.LibraryPlaceBottom,
		Label:     "Teferi's Puzzle Box — put the cards in your hand on the bottom of your library in any order",
		ThenTail:  then,
	}.Apply(ctx)
}

// teferisPuzzleBoxDraw is "then draws that many cards": the player whose
// step it is draws as many as were put on the bottom. An on-disk key:
// never renamed, never reused.
var teferisPuzzleBoxDraw = effects.CardTail("teferis-puzzle-box/draw-that-many", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Tail.N <= 0 {
		return nil
	}
	return effects.DrawCards{Player: o.Tail.Player, N: o.Tail.N}.Apply(ctx)
})
