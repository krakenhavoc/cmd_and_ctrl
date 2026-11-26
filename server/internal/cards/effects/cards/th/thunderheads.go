package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunderheads — Instant {2}{U} (Duel Decks: Izzet vs. Golgari):
//
//	"Replicate {2}{U} (When you cast this spell, copy it for each time you
//	 paid its replicate cost.)
//	 Create a 3/3 blue Weird creature token with defender and flying. Exile it
//	 at the beginning of the next end step."
//
// Pyromatics' replicate (CR 702.56a), one token a copy. Each token is exiled by
// its own delayed trigger at the next end step (CR 603.7), any player's, as
// Rakdos Guildmage's Goblin is: one made during an end step goes at the next
// turn's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "9c5bc3a7-42fe-4f8f-b799-ece7dd98ea4d",
		Name:          "Thunderheads",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
		OptionalCosts: []game.AdditionalCost{effects.ReplicateMana("{2}{U}", 20)},
		Triggered:     []game.TriggeredAbility{effects.Replicate()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			made, err := ctx.Game.CreateTokensForEffect(ctx.Controller(),
				ctx.TextToken(effects.TokenCard("3/3 blue Weird with defender and flying")), 1, game.TokenEntryOptions{})
			if err != nil || len(made) == 0 {
				return err
			}
			return effects.ScheduleDelayedTrigger{
				At:         game.StepEnd,
				Controller: ctx.Controller(),
				Label:      "Thunderheads — exile the Weird token",
				Cards:      made,
				Body:       effects.ExileListedCardsBody,
			}.Apply(ctx)
		},
	})
}
