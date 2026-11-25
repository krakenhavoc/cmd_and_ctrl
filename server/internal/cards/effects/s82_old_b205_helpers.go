package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b205_helpers.go — Battlebond's "for each player, choose friend or
// foe" (Regna's Sanction, Khorvath's Fury; #5529). The controller answers for
// every seated player, themselves first and then in turn order; once every
// answer is in, `done` gets the friends and the foes in that order.

// NewS82OldB205FriendOrFoe registers the answer continuation under key (an
// on-disk identity, never renamed) and returns the function that starts the
// questions for the resolving ability. Call it from an init, because the
// continuation asks the next player with itself.
func NewS82OldB205FriendOrFoe(key, card string, done func(ctx *Context, friends, foes []uuid.UUID) error) func(ctx *Context) error {
	var then game.OptionPickThen
	var ask func(ctx *Context, seats []uuid.UUID, answers []string) error
	finish := func(ctx *Context, answers []string) error {
		var friends, foes []uuid.UUID
		for _, a := range answers {
			if len(a) < 3 {
				continue
			}
			id, err := uuid.Parse(a[2:])
			if err != nil {
				continue
			}
			if a[0] == 'F' {
				friends = append(friends, id)
			} else {
				foes = append(foes, id)
			}
		}
		return done(ctx, friends, foes)
	}
	ask = func(ctx *Context, seats []uuid.UUID, answers []string) error {
		for len(seats) > 0 {
			p := ctx.Game.PlayerByIDForEffect(seats[0])
			if p == nil || p.Eliminated {
				seats = seats[1:]
				continue
			}
			carry := append([]uuid.UUID{seats[0]}, seats[1:]...)
			return PickOption{
				Question: card + " — is " + p.Name + " a friend or a foe?",
				Options:  []game.ChoiceOption{{Label: "Friend"}, {Label: "Foe"}},
				ThenKey:  then,
				Carry:    carry,
				Words:    answers,
				WithItem: true,
			}.Apply(ctx)
		}
		return finish(ctx, answers)
	}
	then = OptionPickThen(key, func(ctx *Context, r game.OptionPicked) error {
		answers := append([]string(nil), r.Words...)
		if len(r.Carry) == 0 {
			return finish(ctx, answers)
		}
		who, rest := r.Carry[0], r.Carry[1:]
		// Unanswered (the chooser left the game): record nothing, carry on.
		switch r.Index {
		case 0:
			answers = append(answers, "F:"+who.String())
		case 1:
			answers = append(answers, "O:"+who.String())
		}
		return ask(ctx, rest, answers)
	})
	return func(ctx *Context) error {
		return ask(ctx, SeatsFromController(ctx.Game, ctx.Controller()), nil)
	}
}
