package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// each_player_may_wheel.go — "each player may discard their hand and draw
// seven cards" (Ruin Grinder, Arc of Fortune, Raphael's Technique).
//
// Every player chooses before anything happens (CR 101.4, CR 608.2e): each
// is asked in APNAP order, eliminated seats skipped, with the yes-sayers so
// far riding the prompt (Turtles in Time's collector). After the last
// answer, every player who said yes discards their hand, and only then does
// each of them draw seven. So nobody answers after seeing another player's
// new hand, and no draw happens between two discards: a discard payoff
// (a madness offer, Library of Leng's replacement) is asked before the
// draws, which wait for it (CR 608.2c, #4155).
//
// The question is the card's own: each card registers its answer under its
// own on-disk key with newMayWheel. The discard-then-draw steps are shared.

// mayWheel is one card's "each player may discard their hand and draw
// seven cards".
type mayWheel struct {
	question string
	answer   game.ConfirmThen
	// draws is how many cards each yes-sayer draws (seven unless a card says
	// otherwise: Will of the Jeskai's five), and steps the wheel that draws
	// them.
	draws int
	steps ThenSteps
}

// newMayWheel registers a card's wheel. answerKey is an on-disk key: never
// renamed, never reused. Call it once, from a package-level var.
func newMayWheel(answerKey, question string) *mayWheel {
	return newMayWheelDrawing(answerKey, question, 7, mayWheelSteps)
}

// newMayWheelDrawing is newMayWheel for a card that draws a number other than
// seven: `steps` is the wheel that draws that many (see mayWheelStepsFor).
func newMayWheelDrawing(answerKey, question string, draws int, steps ThenSteps) *mayWheel {
	w := &mayWheel{question: question, draws: draws, steps: steps}
	// The prompt carries the yes-sayers (the first Word of them), then
	// the players still to ask.
	w.answer = MayAnswer(answerKey, func(ctx *Context, r game.ConfirmAnswer) error {
		n := min(wordInt(r), len(r.Carry))
		yes, rest := append([]uuid.UUID(nil), r.Carry[:n]...), r.Carry[n:]
		if r.Accepted {
			yes = append(yes, r.Chooser)
		}
		return w.ask(ctx, rest, yes)
	})
	return w
}

// Start asks every player, the active player first.
func (w *mayWheel) Start(ctx *Context) error {
	return w.ask(ctx, apnapPlayers(ctx.Game), nil)
}

// ask asks the next player in order; after the last answer, the yes-sayers
// wheel together.
func (w *mayWheel) ask(ctx *Context, order, yes []uuid.UUID) error {
	if len(order) == 0 {
		if len(yes) == 0 {
			return nil
		}
		ctx.RecordStep(mayWheelYesKey, yes)
		return w.steps.Run(ctx)
	}
	who, rest := order[0], order[1:]
	if p := ctx.Game.PlayerByIDForEffect(who); p == nil || p.Eliminated {
		return w.ask(ctx, rest, yes)
	}
	return MayChoice{
		Player:   who,
		Question: w.question,
		Draws:    w.draws,
		Then:     w.answer,
		Carry:    append(append([]uuid.UUID(nil), yes...), rest...),
		Word:     strconv.Itoa(len(yes)),
	}.Apply(ctx)
}

// mayWheelYesKey is the step record holding the players who said yes.
const mayWheelYesKey = "each-player-may-wheel-yes"

// mayWheelSteps is the yes-sayers' wheel: every hand is discarded, then
// each of them draws seven, after any question a discard asked.
var mayWheelSteps = mayWheelStepsFor("then-steps/each-player-may-wheel", 7)

// mayWheelStepsFor registers a wheel that draws `draws` cards under key (an
// on-disk key: never renamed, never reused).
func mayWheelStepsFor(key string, draws int) ThenSteps {
	return RegisterThenSteps(key,
		Instruction(func(ctx *Context) error {
			for _, id := range ctx.StepRecord(mayWheelYesKey).IDs {
				if _, err := discardWholeHand(ctx.Game, id); err != nil {
					return err
				}
			}
			return nil
		}),
		Instruction(func(ctx *Context) error {
			for _, id := range ctx.StepRecord(mayWheelYesKey).IDs {
				if p := ctx.Game.PlayerByIDForEffect(id); p == nil || p.Eliminated {
					continue
				}
				if err := ctx.Game.DrawNForEffect(id, draws); err != nil {
					return err
				}
			}
			return nil
		}),
	)
}
