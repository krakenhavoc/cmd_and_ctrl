package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Court of Ambition — Enchantment {2}{B}{B}:
//
//	"When this enchantment enters, you become the monarch.
//	 At the beginning of your upkeep, each opponent loses 3 life unless
//	 they discard a card. If you're the monarch, instead each opponent
//	 loses 6 life unless they discard two cards."
//
// The black Court (#1722). The "unless" is each opponent's own choice,
// made in APNAP order (CR 101.4) so a later player sees what the earlier
// ones chose — Torment of Hailfire's sequential PickOption chain, one
// question per opponent, each asked only once the previous one has been
// answered and its discard has happened.
//
// Both numbers are read at resolution (the monarch clause is part of
// the effect), once, before the first question: the crown cannot move
// mid-resolution, so every opponent faces the same price.
//
// A player who cannot discard the full count cannot take the "unless"
// (CR 118.3 — a cost can't be paid without the resources), so they are
// not asked and simply lose the life: an empty hand against the small
// half, fewer than two cards against the monarch half. Offering a
// discard the player cannot complete would be offering them a way out
// the card does not print.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "652e71a9-e46f-41b3-8695-76b3606b1955",
		Name:         "Court of Ambition",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Court of Ambition"),
			AtYourUpkeep("Court of Ambition — each opponent loses 3 life unless they discard a card, or 6 unless they discard two if you're the monarch",
				courtOfAmbitionUpkeep),
		},
	})
}

// courtOfAmbitionUpkeep fixes the price and starts the run of questions.
func courtOfAmbitionUpkeep(g *game.Game, item *game.StackItem) error {
	life, discard := 3, 1
	if YoureTheMonarch(g, item.Controller) {
		life, discard = 6, 2
	}
	var victims []uuid.UUID
	for _, id := range apnapPlayers(g) {
		if id != item.Controller {
			victims = append(victims, id)
		}
	}
	return courtOfAmbitionStep(NewContext(g, item), victims, life, discard)
}

// courtOfAmbitionStep asks the next opponent still owed a question.
// Package-level over frozen scalars, so an undo that replays an answer
// resolves it against the restored game.
//
// Caller holds g.mu.
func courtOfAmbitionStep(ctx *Context, remaining []uuid.UUID, life, discard int) error {
	for len(remaining) > 0 {
		victim, rest := remaining[0], remaining[1:]
		p := ctx.Game.PlayerByIDForEffect(victim)
		if p == nil || p.Eliminated {
			remaining = rest
			continue
		}
		if p.Hand == nil || p.Hand.Size() < discard {
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -life); err != nil {
				return err
			}
			remaining = rest
			continue
		}
		cards := "a card"
		if discard == 2 {
			cards = "two cards"
		}
		return PickOption{
			Player:   victim,
			Question: "Court of Ambition — lose " + strconv.Itoa(life) + " life, or discard " + cards + "?",
			Options: []game.ChoiceOption{
				{Label: "Lose " + strconv.Itoa(life) + " life", LifeCost: life},
				{Label: "Discard " + cards},
			},
			Then: courtOfAmbitionAnswered(victim, rest, life, discard),
		}.Apply(ctx)
	}
	return nil
}

// courtOfAmbitionAnswered runs the chosen branch and then the next
// question.
func courtOfAmbitionAnswered(victim uuid.UUID, rest []uuid.UUID, life, discard int) func(ctx *Context, index int) error {
	return func(ctx *Context, index int) error {
		switch index {
		case 1:
			item := ctx.Item
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player:   victim,
				Source:   ctx.Source(),
				N:        discard,
				Question: "Court of Ambition — discard",
				Then: func(g *game.Game, _ uuid.UUID, _ []uuid.UUID) error {
					return courtOfAmbitionStep(NewContext(g, item), rest, life, discard)
				},
			})
			return nil
		case 0:
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -life); err != nil {
				return err
			}
		}
		// index -1: the player left before the question could be
		// asked (CR 800.4a); the rest of the run still happens.
		return courtOfAmbitionStep(ctx, rest, life, discard)
	}
}
