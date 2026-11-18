package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zndrsplt's Judgment — Sorcery {4}{U} (S82 reprint pass):
//
//	"For each player, choose friend or foe. Each friend creates a token
//	 that's a copy of a creature they control. Each foe returns a creature
//	 they control to its owner's hand."
//
// Two stages, each a chain of prompts so a table waiting on any of them is
// a restore point. First the CASTER is asked, one player at a time and
// starting with themselves, whether that player is a friend or a foe; the
// answers so far ride the prompt as words ("F:<id>" / "O:<id>"). Then the
// printed order is followed: each friend in turn picks a creature they
// control and gets a token that is a copy of it, then each foe picks a
// creature they control and returns it to its owner's hand. A player with no
// creature is not asked and does nothing; a creature that left before its
// picker answered is skipped. The remaining players ride the prompt as its
// carry, friends first, a nil id separating them from the foes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7aacfc42-be6c-46a8-b474-f43a1b6b2ef8",
		Name:         "Zndrsplt's Judgment",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return zndrspltAsk(ctx, seatsFromController(ctx.Game, item.Controller), nil)
		},
	})
}

// zndrspltAsk asks the caster about seats[0], carrying the rest and the
// answers so far. With nobody left to ask it moves on to the pick stage.
func zndrspltAsk(ctx *Context, seats []uuid.UUID, answers []string) error {
	g := ctx.Game
	for len(seats) > 0 {
		p := g.PlayerByIDForEffect(seats[0])
		if p == nil || p.Eliminated {
			seats = seats[1:]
			continue
		}
		carry := append([]uuid.UUID{seats[0]}, seats[1:]...)
		return PickOption{
			Question: "Zndrsplt's Judgment — is " + p.Name + " a friend or a foe?",
			Options:  []game.ChoiceOption{{Label: "Friend"}, {Label: "Foe"}},
			ThenKey:  zndrspltFriendOrFoe,
			Carry:    carry,
			Words:    answers,
			WithItem: true,
		}.Apply(ctx)
	}
	return zndrspltPlan(ctx, answers)
}

// zndrspltFriendOrFoe records the answer for the player being asked (the
// first of the carry) and asks about the next. An on-disk key, registered in
// init because its body asks the next seat with it.
var zndrspltFriendOrFoe game.OptionPickThen

func init() {
	zndrspltFriendOrFoe = OptionPickThen("option-pick/zndrsplts-judgment-friend-or-foe", func(ctx *Context, r game.OptionPicked) error {
		answers := append([]string(nil), r.Words...)
		if len(r.Carry) == 0 {
			return zndrspltPlan(ctx, answers)
		}
		who, rest := r.Carry[0], r.Carry[1:]
		// Unanswered (the caster left the game): record nothing, carry on.
		switch r.Index {
		case 0:
			answers = append(answers, "F:"+who.String())
		case 1:
			answers = append(answers, "O:"+who.String())
		}
		return zndrspltAsk(ctx, rest, answers)
	})
}

// zndrspltPlan turns the answers into the carry of the pick stage, friends
// first, then a nil id, then foes, and starts it.
func zndrspltPlan(ctx *Context, answers []string) error {
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
	plan := append(append(friends, uuid.Nil), foes...)
	return zndrspltStep(ctx.Game, ctx.Source(), plan)
}

// zndrspltStep asks the next player in the plan to pick a creature. The
// plan is friends, a nil separator, then foes; a nil separator still in the
// rest means the player being asked is a friend.
func zndrspltStep(g *game.Game, source uuid.UUID, plan []uuid.UUID) error {
	for len(plan) > 0 {
		if plan[0] == uuid.Nil {
			plan = plan[1:]
			continue
		}
		who, rest := plan[0], plan[1:]
		p := g.PlayerByIDForEffect(who)
		var creatures []uuid.UUID
		for _, c := range g.Battlefield.Cards {
			if c.Controller == who && c.IsCreature() {
				creatures = append(creatures, c.InstanceID)
			}
		}
		if p == nil || p.Eliminated || len(creatures) == 0 {
			plan = rest
			continue
		}
		word, question := "foe", "Zndrsplt's Judgment — return a creature you control to its owner's hand"
		if slices.Contains(rest, uuid.Nil) {
			word, question = "friend", "Zndrsplt's Judgment — choose a creature you control to copy"
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:   who,
			Source:    source,
			Question:  question,
			Cards:     creatures,
			Min:       1,
			Max:       1,
			LiveZone:  game.ZoneBattlefield,
			ThenKey:   zndrspltCreaturePicked,
			Carry:     rest,
			CarryWord: word,
		})
		return nil
	}
	return nil
}

// zndrspltCreaturePicked copies (a friend) or returns (a foe) the creature
// picked, then asks the next player in the plan. An on-disk key.
var zndrspltCreaturePicked game.ChooseCardsThen

func init() {
	zndrspltCreaturePicked = game.RegisterChooseCardsThen("choose-cards/zndrsplts-judgment-creature", func(g *game.Game, r game.CardsChosen) error {
		ctx := NewContext(g, &game.StackItem{
			Kind:         game.StackItemSpell,
			Controller:   r.Chooser,
			Owner:        r.Chooser,
			SourceCardID: r.Source,
		})
		if len(r.Picked) > 0 {
			id := r.Picked[0]
			if c, ok := g.LookupCardForEffect(id); ok && g.Battlefield.Contains(id) && c.Controller == r.Chooser {
				var err error
				if r.Word == "friend" {
					err = CreateTokenCopy{Controller: r.Chooser, Copy: id, N: 1}.Apply(ctx)
				} else {
					err = BounceToHand{Target: id}.Apply(ctx)
				}
				if err != nil {
					return err
				}
			}
		}
		return zndrspltStep(g, r.Source, r.Carry)
	})
}
