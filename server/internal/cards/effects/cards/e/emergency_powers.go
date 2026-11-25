package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emergency Powers — Instant {5}{W}{U} (Ravnica Allegiance):
//
//	"Each player shuffles their hand and graveyard into their library, then
//	 draws seven cards. Exile Emergency Powers.
//	 Addendum — If you cast this spell during your main phase, you may put a
//	 permanent card with mana value 7 or less from your hand onto the
//	 battlefield."
//
// Echo of Eons' shuffle-and-draw chain, which can pause on a commander
// prompt, so it is keyed (emergencyPowers); the spell is exiled and the
// addendum's optional put follows the draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9127ac59-4c7b-4652-b1e5-d82af47f2717",
		Name:         "Emergency Powers",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve:    emergencyPowers,
	})
}

// emergencyPowers is Emergency Powers' body: each player shuffles their
// hand and graveyard into their library and draws seven, the spell is exiled,
// and, if it was cast in its controller's main phase, the controller may put
// a permanent card with mana value 7 or less from their hand onto the
// battlefield. The tuck can pause on a CR 903.9b prompt, so the chain is keyed
// (Echo of Eons' shape) and carries the controller ahead of the roster.
func emergencyPowers(item *game.StackItem, ctx *effects.Context) error {
	players := effects.TablePlayers(ctx)
	// Exile the spell first: on the stack it is in nobody's graveyard, so the
	// order against the shuffles cannot be seen.
	if !effects.SourceIsNewObject(ctx.Game, item) {
		if err := ctx.Game.ExileCardForEffect(item.SourceCardID); err != nil {
			return err
		}
	}
	return emergencyShuffleEach(ctx.Game, ctx.Controller(), item.SourceCardID, players, players)
}

// emergencyShuffleEach tucks the head player's hand and graveyard into
// their library, then continues with the rest; once all have shuffled, every
// player draws seven and the addendum is offered.
func emergencyShuffleEach(g *game.Game, controller, source uuid.UUID, remaining, all []uuid.UUID) error {
	if len(remaining) == 0 {
		return emergencyFinish(g, controller, source, all)
	}
	id, rest := remaining[0], remaining[1:]
	ids := append(effects.AllHandCardIDs(g, id), effects.EchoOfEonsGraveyardCardIDs(g, id)...)
	carry := append(append([]uuid.UUID{controller}, all...), rest...)
	return g.TuckCardsToLibraryKeyedForEffect(ids, game.TuckOptions{},
		game.KeyedTail{Ref: emergencyPowersNext, Player: id, Source: source, Carry: carry, N: len(all)})
}

// emergencyFinish is the draw and the addendum.
func emergencyFinish(g *game.Game, controller, source uuid.UUID, all []uuid.UUID) error {
	if err := effects.EchoOfEonsDrawEach(g, all); err != nil {
		return err
	}
	if !effects.B14InYourMainPhase(g, controller) {
		return nil
	}
	return effects.PutFromHandOntoBattlefield{
		Player:   controller,
		Match:    effects.ManaValueLE(7),
		Optional: true,
		Label:    "Emergency Powers — you may put a permanent card with mana value 7 or less from your hand onto the battlefield",
	}.Apply(effects.KeyedContext(g, controller, source))
}

var emergencyPowersNext game.TailRef

func init() {
	emergencyPowersNext = game.RegisterTail("then/old-b179-emergency-powers-next", func(g *game.Game, o game.TailOutcome) error {
		if err := g.ShuffleLibraryForEffect(o.Tail.Player); err != nil {
			return err
		}
		if len(o.Tail.Carry) == 0 {
			return nil
		}
		controller, carry := o.Tail.Carry[0], o.Tail.Carry[1:]
		n := min(max(o.Tail.N, 0), len(carry))
		return emergencyShuffleEach(g, controller, o.Tail.Source, carry[n:], carry[:n])
	})
}
