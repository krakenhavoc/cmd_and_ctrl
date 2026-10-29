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
	courtOfAmbitionLoss = RegisterThenSteps("then-steps/court-of-ambition-loss", courtOfAmbitionLossSteps()...)
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
			return courtOfAmbitionLoseThenNext(ctx, victim, rest, life)
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
			ThenKey: courtOfAmbitionPrices[life].answered,
			Carry:   append([]uuid.UUID{victim}, rest...),
		}.Apply(ctx)
	}
	return nil
}

// courtOfAmbitionPrice is the keyed continuations of one price of the card
// (#4253): the life lost and, with it, the cards discarded instead (a third
// of the life, 1 for 3 and 2 for 6). The price is part of the key, because
// a registered body captures nothing and an option pick carries IDs only.
type courtOfAmbitionPrice struct {
	// answered is the opponent's choice between losing the life and
	// discarding; its Carry is the opponent, then the opponents still owed
	// a question.
	answered game.OptionPickThen
	// discarded is the discard behind the choice; its Carry is the
	// effect's controller, then the opponents still owed a question.
	discarded game.DiscardThen
}

// courtOfAmbitionPrices is registered in init, because each answer asks the
// next opponent and so names itself. The keys are on-disk identities: never
// renamed, never reused.
var courtOfAmbitionPrices = map[int]courtOfAmbitionPrice{}

func init() {
	for _, life := range []int{3, 6} {
		discard := life / 3
		suffix := strconv.Itoa(life)
		price := courtOfAmbitionPrice{}
		price.discarded = game.RegisterDiscardThen("choose-cards/court-of-ambition-discarded-"+suffix,
			func(g *game.Game, r game.DiscardsDone) error {
				if len(r.Carry) == 0 {
					return nil
				}
				return courtOfAmbitionStep(keyedContext(g, r.Carry[0], r.Source), r.Carry[1:], life, discard)
			})
		price.answered = OptionPickThen("option-pick/court-of-ambition-answer-"+suffix,
			func(ctx *Context, r game.OptionPicked) error {
				if len(r.Carry) == 0 {
					return nil
				}
				victim, rest := r.Carry[0], r.Carry[1:]
				switch r.Index {
				case 1:
					return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player:   victim,
						Source:   ctx.Source(),
						N:        discard,
						Question: "Court of Ambition — discard",
					}, courtOfAmbitionPrices[life].discarded, append([]uuid.UUID{ctx.Controller()}, rest...))
				case 0:
					return courtOfAmbitionLoseThenNext(ctx, victim, rest, life)
				default:
					// index -1: the player left before the question could be
					// asked (CR 800.4a); the rest of the run still happens.
					return courtOfAmbitionStep(ctx, rest, life, discard)
				}
			})
		courtOfAmbitionPrices[life] = price
	}
}

// courtOfAmbitionLoseThenNext is one opponent's life loss, then the next
// opponent's question, as continuation steps (#4162, CR 608.2c): the
// question waits for a prompt the loss asks (a CR 616 order of two
// life-loss replacements). The record carries the victim, the opponents
// still owed a question, and the life (the discard count is a third of
// it: 1 for 3, 2 for 6).
func courtOfAmbitionLoseThenNext(ctx *Context, victim uuid.UUID, rest []uuid.UUID, life int) error {
	ctx.RecordStep(courtOfAmbitionRecord, append([]uuid.UUID{victim}, rest...), life)
	return courtOfAmbitionLoss.Run(ctx)
}

const courtOfAmbitionRecord = "court-of-ambition/loss"

// courtOfAmbitionLoss is registered in init, because its last step asks
// the next opponent, whose answer runs it again.
var courtOfAmbitionLoss ThenSteps

func courtOfAmbitionLossSteps() []Applier {
	return []Applier{
		Instruction(func(ctx *Context) error {
			r := ctx.StepRecord(courtOfAmbitionRecord)
			if len(r.IDs) == 0 {
				return nil
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.IDs[0], -r.N())
		}),
		Instruction(func(ctx *Context) error {
			r := ctx.StepRecord(courtOfAmbitionRecord)
			if len(r.IDs) == 0 {
				return nil
			}
			return courtOfAmbitionStep(ctx, r.IDs[1:], r.N(), r.N()/3)
		}),
	}
}
