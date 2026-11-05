package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Turtle Tracks — Sorcery {2}{G} (Teenage Mutant Ninja Turtles Commander,
// #5011):
//
//	"Any number of target players may each search their library for a
//	 basic land card, put it onto the battlefield under their control, then
//	 shuffle."
//
// "Any number of target players" is Stonespeaker Crystal's player clause;
// zero targets is a legal announcement. On resolution each target still in
// the game is offered the search in APNAP order (CR 101.4), one library at
// a time through the keyed each-player chain (eachPlayerInTurn), so only
// one search prompt is ever open. Each search is that player's own and
// optional ("may"): a player who declines neither searches nor shuffles,
// and one who searches may still fail to find (CR 701.23b). The land enters
// untapped under the searcher's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4921dac2-4adf-4ac4-b1f4-d53385afd802",
		Name:         "Turtle Tracks",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 1, Tutors: 1},
		Targets:      TargetPlayer("any number of target players").WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var players []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					players = append(players, t.ID)
				}
			}
			return eachPlayerInTurn(ctx.Game, ctx.ItemTail(turtleTracksStep), players, nil, game.KeyedTail{}, turtleTracksSearch)
		},
	})
}

// turtleTracksStep is one player's search in Turtle Tracks' chain. An
// on-disk key: never renamed, never reused.
var turtleTracksStep = registerEachPlayerStep("turtle-tracks/search-step", turtleTracksSearch)

// turtleTracksSearch is one player's optional search for a basic land onto
// the battlefield, ending on the chain's next step.
func turtleTracksSearch(ctx *Context, player uuid.UUID, _ int, step game.KeyedTail) error {
	return SearchLibrary{
		Player:    player,
		Predicate: IsBasicLand,
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		Optional:  true,
		Reason:    "Turtle Tracks — you may search for a basic land",
		ThenTail:  step,
	}.Apply(ctx)
}
