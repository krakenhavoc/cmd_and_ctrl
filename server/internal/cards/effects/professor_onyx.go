package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Professor Onyx — Legendary Planeswalker — Liliana {4}{B}{B}, starting
// loyalty 5:
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	 each opponent loses 2 life and you gain 2 life.
//	 +1: You lose 1 life. Look at the top three cards of your library. Put
//	 one of them into your hand and the rest into your graveyard.
//	 −3: Each opponent sacrifices a creature with the greatest power among
//	 creatures that player controls.
//	 −8: Each opponent may discard a card. If they don't, they lose 3 life.
//	 Repeat this process six more times."
//
// The −8 is seven rounds of the same offer. A round asks each opponent in
// turn order (the controller's left first); the whole run is a queue of
// seats carried as plain data, so a table waiting on any answer is a
// restore point. "May discard a card" is an up-to-one discard prompt: a
// player who discards nothing, including one with an empty hand, loses 3
// life (a life change, no damage).
func init() {
	Register(Spec{
		OracleID:        "9df2e909-ed13-456a-9636-7398732009a9",
		Name:            "Professor Onyx",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		Triggered: []game.TriggeredAbility{
			Magecraft("Professor Onyx — each opponent loses 2 life and you gain 2 life",
				Do(EachOpponentLosesLife{N: 2}, GainLife{Amount: 2})),
		},
		Activated: []ActivatedAbility{
			{
				Label: "+1: You lose 1 life. Look at the top three cards of your library. Put one of them into your hand and the rest into your graveyard.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1); err != nil {
						return err
					}
					player := ctx.Controller()
					return TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 3),
						Max:    1,
						Label:  "Professor Onyx — put one of them into your hand and the rest into your graveyard",
						Key:    professorOnyxTake,
					}.Apply(ctx)
				},
			},
			{
				Label: "−3: Each opponent sacrifices a creature with the greatest power among creatures that player controls.",
				Cost:  LoyaltyCost(-3),
				Effect: func(g *game.Game, item *game.StackItem) error {
					g.PlayersSacrificeForEffect(item.SourceCardID, opponentsInTurnOrder(g, item.Controller),
						sacrificeSpec("a creature with the greatest power among creatures you control", GreatestPowerYouControl()),
						"Professor Onyx — sacrifice a creature with the greatest power")
					return nil
				},
			},
			{
				Label: "−8: Each opponent may discard a card. If they don't, they lose 3 life. Repeat this process six more times.",
				Cost:  LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					opps := opponentsInTurnOrder(g, item.Controller)
					var queue []uuid.UUID
					for round := 0; round < 7; round++ {
						queue = append(queue, opps...)
					}
					return professorOnyxAsk(g, item.SourceCardID, queue)
				},
			},
		},
	})
}

var professorOnyxTake = RegisterTake("choose-cards/professor-onyx-take",
	TakeDef{Then: func(g *game.Game, res TakeFromLibraryResult) error {
		return TakeRestIntoGraveyard(g, res)
	}})

// professorOnyxAsk offers the next seat in the queue its discard, with the
// seats after it as the carry. An empty queue ends the run.
func professorOnyxAsk(g *game.Game, source uuid.UUID, queue []uuid.UUID) error {
	for len(queue) > 0 {
		next, rest := queue[0], queue[1:]
		if p := g.PlayerByIDForEffect(next); p == nil || p.Eliminated {
			queue = rest
			continue
		}
		return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   next,
			Source:   source,
			N:        1,
			UpTo:     true,
			Question: "Professor Onyx — you may discard a card; if you don't, you lose 3 life",
		}, professorOnyxAnswer, rest)
	}
	return nil
}

// professorOnyxAnswer is one opponent's answer: no discard costs 3 life,
// then the run goes on with the next seat. Carry is the seats still to ask.
// Registered from init: the answer asks the next seat, which names the
// answer again.
var professorOnyxAnswer game.DiscardThen

func init() {
	professorOnyxAnswer = game.RegisterDiscardThen("choose-cards/professor-onyx-discard-or-lose", func(g *game.Game, r game.DiscardsDone) error {
		if !r.Discarded.Discarded(r.Player) {
			if err := g.ChangePlayerLifeForEffect(r.Source, r.Player, -3); err != nil {
				return err
			}
		}
		return professorOnyxAsk(g, r.Source, r.Carry)
	})
}
