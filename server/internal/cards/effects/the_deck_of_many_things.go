package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Deck of Many Things — Legendary Artifact {5} (Adventures in the
// Forgotten Realms, #5681):
//
//	"{2}, {T}: Roll a d20 and subtract the number of cards in your hand.
//	 If the result is 0 or less, discard your hand.
//	 1—9 | Return a card at random from your graveyard to your hand.
//	 10—19 | Draw two cards.
//	 20 | Put a creature card from any graveyard onto the battlefield
//	 under your control. When that creature dies, its owner loses the
//	 game."
//
// The hand, counted as the ability resolves, is the roll's modifier (CR
// 706.2, RollDice.Modifier with a negative number), so the die is
// announced with its natural result and the hand beside it: a die showing
// 20 with three cards in hand is a natural 20 and a result of 17. The
// result picks the row (CR 706.3a): 0 or less discards the hand, and the
// 20 row needs a result of exactly 20.
//
// The 1—9 card is picked from the keyed random stream (ADR 0054). The 20
// row's creature card is chosen as the row is carried out, from any
// graveyard, and is not a target. Once it has entered, a delayed
// triggered ability about that permanent as the object it is (CR 603.7c)
// watches for it dying, for as long as it is that object (no "this turn"
// is printed), and makes its owner lose the game, whoever controls it
// then and whoever controls the Deck.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "91df4cf5-d7ec-4fcd-87ed-e075ef6ceba9",
		Name:         "The Deck of Many Things",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Roll a d20 and subtract the number of cards in your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				hand := 0
				if p := g.PlayerByIDForEffect(ctx.Controller()); p != nil && p.Hand != nil {
					hand = p.Hand.Size()
				}
				return RollDice{Sides: 20, N: 1, Modifier: -hand, Then: deckOfManyThingsResult}.Apply(ctx)
			},
		}},
	})
}

// deckOfManyThingsResult is the results table, after the d20. An on-disk
// key: never renamed, never reused.
var deckOfManyThingsResult = CardTail("dice/the-deck-of-many-things-result", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Rolls) == 0 {
		return nil
	}
	you := ctx.Controller()
	switch result := DiceResult(o); {
	case result <= 0:
		_, err := discardWholeHand(ctx.Game, you)
		return err
	case result <= 9:
		picked := randomPick(ctx, graveyardIDs(ctx.Game, you, nil), 1)
		if len(picked) == 0 {
			return nil
		}
		return ReturnFromGraveyard{Target: picked[0], Dest: game.ZoneHand}.Apply(ctx)
	case result <= 19:
		return DrawCards{Player: you, N: 2}.Apply(ctx)
	case result == 20:
		candidates := allGraveyardsCardIDs(ctx, game.Card.IsCreature)
		if len(candidates) == 0 {
			return nil
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  you,
			Source:   ctx.Source(),
			Question: "The Deck of Many Things — put a creature card from any graveyard onto the battlefield under your control (when it dies, its owner loses the game)",
			Cards:    candidates,
			Min:      1,
			Max:      1,
			ThenKey:  deckOfManyThingsPicked,
			Asked:    game.NewAskedItem(ctx.Item),
		})
	}
	return nil
})

// deckOfManyThingsPicked is the chosen creature card entering under the
// activator's control. An on-disk key: never renamed, never reused.
var deckOfManyThingsPicked = askedPick("choose-cards/the-deck-of-many-things-reanimate", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ReturnFromGraveyardTogether{
		Targets:    []uuid.UUID{r.Picked[0]},
		Controller: ctx.Controller(),
		ThenTail:   deckOfManyThingsCurse,
	}.Apply(ctx)
})

// deckOfManyThingsCurse is "when that creature dies, its owner loses the
// game", created once the creature has entered: a one-shot delayed
// trigger about the permanent as the object it is now, with no end. An
// on-disk key: never renamed, never reused.
var deckOfManyThingsCurse = TogetherTail("the-deck-of-many-things/curse", func(ctx *Context, entered []uuid.UUID, _ int) error {
	for _, id := range entered {
		c, ok := battlefieldCardForEffect(ctx.Game, id)
		if !ok {
			continue
		}
		forever := game.IndefiniteDuration()
		ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
			Controller:   ctx.Controller(),
			SourceCardID: ctx.Source(),
			Label:        "The Deck of Many Things — its owner loses the game",
			On:           []game.EventKind{game.EventLTB},
			Condition:    thatCreatureDiedCondition,
			CondParams:   game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
			Body:         deckOfManyThingsOwnerLosesBody,
			Duration:     &forever,
		})
	}
	return nil
})

// deckOfManyThingsOwnerLosesBody is the fired trigger: the owner of the
// creature that died loses the game.
var deckOfManyThingsOwnerLosesBody = game.SimpleDelayedBody("the-deck-of-many-things/owner-loses", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	obj, ok := payloadObject(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(obj.ID)
	if !ok || c.Owner == uuid.Nil {
		return nil
	}
	return LoseTheGame{Player: c.Owner}.Apply(ctx)
})
