package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gleaming Splendor — Enchantment {1}{W}:
//
//	"Whenever an opponent draws their second card each turn, you create
//	 a Treasure token.
//	 {2}{W}: Two target players each draw a card."
//
// The trigger is Faerie Mastermind's (opponentDrewTheirSecondCardThisTurn:
// the opponent's own draw tally reaches exactly two, so it fires once per
// opponent per turn, and a draw from the Splendor's own ability counts).
// The ability needs exactly two different players, which the target
// clause (two, no more, no fewer) enforces; the draws are no-choice, so
// each player draws in turn order (CR 101.4).
//
// No simplification. The Hobbit (#4454).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c01aeaa5-1d3b-4493-9575-30175dcd780d",
		Name:         "Gleaming Splendor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.OpponentDrewTheirSecondCardThisTurn(ev, source, g)
			}, "Gleaming Splendor — create a Treasure", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: Two target players each draw a card.",
			Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 1}),
			Cost:    effects.ManaCost("{2}{W}"),
			Targets: effects.TargetPlayer("two target players").WithCount(2, 2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				var players []uuid.UUID
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						players = append(players, t.ID)
					}
				}
				for _, p := range players {
					if err := (effects.DrawCards{Player: p, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
