package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// extra_turns.go — the card side of CR 500.7 extra turns (ADR 0059
// Decision 5, #753).
//
// "Take an extra turn after this one" is TakeExtraTurn and nothing
// else: the engine queues the turn, takes the most recently created
// first, resumes normal rotation afterwards and drops the turn of a
// player who has left (game/extra_turns.go). A card never touches the
// turn cursor.
//
// "At the beginning of that turn's end step, you lose the game" (Final
// Fortune, Last Chance, Warrior's Oath) is TakeExtraTurnThenLose: the
// same turn, plus a delayed trigger BOUND to it
// (game.DelayedTrigger.OnExtraTurn), so it fires in the extra turn's end
// step and in no other — and never, if the extra turn is skipped.

// TakeExtraTurn is "[target player] take[s] N extra turns after this
// one" (CR 500.7).
type TakeExtraTurn struct {
	// Player takes the turns. Zero means the resolving item's
	// controller ("take an extra turn").
	Player uuid.UUID
	// N is how many; zero means one. Time Stretch and Teferi, Master of
	// Time's −10 say two.
	N int
	// SkipSteps are steps of "that turn" the turn skips: Savor the
	// Moment's "Skip the untap step of that turn" is
	// SkipSteps: []game.Step{game.StepUntap} (ADR 0188 Decision 3a). Each
	// is spent by the first matching step of the turn, so an untap step
	// in a beginning phase added later in the same turn still happens.
	// The turn that is skipped takes the record with it.
	SkipSteps []game.Step
	// DuringThatTurn are rules effects that last during "that turn":
	// Alchemist's Gambit's "During that turn, damage can't be prevented"
	// is DuringThatTurn: []game.Mod{{Kind: game.ModDamageCantBePrevented}}
	// (ADR 0188 Decision 3b). They are game-wide (ScopeGame) rules
	// gates, queued with the turn and registered as it begins, so they
	// cover none of the turn the spell resolved in and end at the extra
	// turn's cleanup step. Rules only: an effect on objects would fix its
	// set when it begins (CR 611.2c).
	DuringThatTurn []game.Mod
	// Label names the effect in the log and on the record. Defaults to
	// "Extra turn". Used only with SkipSteps or DuringThatTurn.
	Label string
}

// Apply implements Primitive. It queues and never fails: a player who
// has left the game takes nothing (CR 800.4a).
func (x TakeExtraTurn) Apply(ctx *Context) error {
	x.refs(ctx)
	return nil
}

// refs queues the turns and returns their refs, the one taken first
// first — what a card that binds something to "that turn" needs. "That
// turn" is the first one taken, so SkipSteps and DuringThatTurn bind to
// refs[0]: the cards that print them take one turn.
func (x TakeExtraTurn) refs(ctx *Context) []int {
	player := x.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	n := x.N
	if n == 0 {
		n = 1
	}
	refs := ctx.Game.TakeExtraTurnsForEffect(player, ctx.Source(), n)
	if len(refs) == 0 {
		return refs
	}
	label := x.Label
	if label == "" {
		label = "Extra turn"
	}
	for _, step := range x.SkipSteps {
		ctx.Game.SkipStepOfExtraTurnForEffect(refs[0], step, ctx.Source(), label)
	}
	if len(x.DuringThatTurn) > 0 {
		ctx.Game.DuringExtraTurnRuleEffectForEffect(refs[0], ctx.Source(), game.ScopeGame,
			uuid.Nil, x.DuringThatTurn, label)
	}
	return refs
}

// TakeExtraTurnThenLose is Final Fortune's sentence: "Take an extra
// turn after this one. At the beginning of that turn's end step, you
// lose the game." The loss is a delayed triggered ability (CR 603.7)
// controlled by the spell's controller and bound to the extra turn, so:
//
//   - it goes on the stack in that turn's end step and can be answered
//     (a Stifle saves the player);
//   - it does not fire in the end step of the turn the spell resolved
//     in, which is the one an unbound "next end step" would pick;
//   - if the extra turn never begins, or ends before its end step, it
//     is swept and the player does not lose — the card's 2004-10-04
//     ruling.
type TakeExtraTurnThenLose struct {
	// Label is the delayed trigger's stack label:
	// "Final Fortune — you lose the game".
	Label string
	// Turn is the extra turn taken, for a card that gives it more than
	// the loss: Alchemist's Gambit's "During that turn, damage can't be
	// prevented" (TakeExtraTurn.DuringThatTurn). The zero value is a
	// plain extra turn for the spell's controller.
	Turn TakeExtraTurn
}

// Apply implements Primitive.
func (x TakeExtraTurnThenLose) Apply(ctx *Context) error {
	refs := x.Turn.refs(ctx)
	if len(refs) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:          game.StepEnd,
		Label:       x.Label,
		Body:        loseTheGameAtThatTurnsEndBody,
		OnExtraTurn: refs[0],
	}.Apply(ctx)
}

// extraTurnThenLoseAtItsEndStep is the resolving body of the three
// cards that print Final Fortune's sentence.
func extraTurnThenLoseAtItsEndStep(label string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		return TakeExtraTurnThenLose{Label: label}.Apply(ctx)
	}
}

// youTakeAnExtraTurn is the resolving body of a spell that says only
// "Take an extra turn after this one" (Temporal Manipulation, Capture
// of Jingzhou).
func youTakeAnExtraTurn(_ *game.StackItem, ctx *Context) error {
	return TakeExtraTurn{}.Apply(ctx)
}

// youTakeAnExtraTurnEffect is the same sentence as an activated
// ability's effect (Magistrate's Scepter, Avatar Kuruk).
func youTakeAnExtraTurnEffect(g *game.Game, item *game.StackItem) error {
	return TakeExtraTurn{}.Apply(NewContext(g, item))
}

// loseTheGameAtThatTurn is the delayed trigger's body: its controller
// loses the game (ADR 0057). The loser controls the resolving trigger,
// so LoseTheGame returns game.ErrStopResolution and the rotation waits
// for the resolution bookend's SBA pass (ADR 0059 Decision 6).
func loseTheGameAtThatTurn(g *game.Game, item *game.StackItem) error {
	return LoseTheGame{}.Apply(NewContext(g, item))
}
