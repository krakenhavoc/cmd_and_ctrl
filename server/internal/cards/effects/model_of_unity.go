package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Model of Unity — Artifact {3}:
//
//	"Whenever players finish voting, you and each opponent who voted for
//	 a choice you voted for may scry 2.
//	 {T}: Add one mana of any color."
//
// ADR 0146's amendment of 2026-10-13 (finish_voting.go). The players
// affected are fixed when the ability resolves, from the finished
// vote's ballots: you (whether or not you voted), then each opponent who
// shares a choice with you, in seat order. Each is asked "scry 2?" in
// turn and a yes opens their scry before the next player is asked, so
// one player's answer never waits on another's cards. Every prompt is
// an ordinary confirm and scry addressed to its player, so a bot seat
// answers its own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d7b69fd7-e575-4009-8549-f576249ba641",
		Name:         "Model of Unity",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			WheneverPlayersFinishVoting("Model of Unity — you and each opponent who voted for a choice you voted for may scry 2",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					players := append([]uuid.UUID{ctx.Controller()}, OpponentsWhoVotedForAChoiceYouVotedFor(ctx)...)
					return mayScryInTurn(ctx, players, 2)
				}),
		},
	})
}

// mayScryInTurn asks each player in turn "you may scry n", opening a
// yes-answer's scry before the next player is asked.
func mayScryInTurn(ctx *Context, players []uuid.UUID, n int) error {
	if len(players) == 0 {
		return nil
	}
	p, rest := players[0], players[1:]
	next := func(ctx *Context) error { return mayScryInTurn(ctx, rest, n) }
	return MayChoice{
		Player:   p,
		Question: "Model of Unity — scry 2?",
		OnYes: func(ctx *Context) error {
			item := ctx.Item
			return Scry{
				Player: p,
				N:      n,
				Then:   func(g *game.Game) error { return next(NewContext(g, item)) },
			}.Apply(ctx)
		},
		OnNo: next,
	}.Apply(ctx)
}
