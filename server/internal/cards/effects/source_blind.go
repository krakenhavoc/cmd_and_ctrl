package effects

import (
	"errors"
	"reflect"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// source_blind.go — which triggered rows are SOURCE-BLIND (#1968,
// ADR 0018's #1968 amendment): their Effect reads nothing of its stack
// item but the controller. Two Soul Wardens' triggers are then one
// effect queued twice, and the CR 603.3b order between them cannot
// change the game, so the drain does not ask for it
// (game.TriggeredAbility.SourceBlind, seatNeedsTriggerOrder).
//
// An Effect is a closure and the engine cannot see what it reads. What
// the registry CAN see is a row whose Effect is a Do of primitive
// values: Do's closure answers a probe with its steps, and each step's
// type says what it reads. So the class is: no Build, no target or
// mode clause, and an Effect that is a Do of primitives on the list in
// sourceBlindStep. Everything else keeps asking. That is the safe
// direction — a row left off the list costs a click, a row wrongly on
// it takes a real choice away.
//
// Adding a primitive to the list needs the argument the amendment makes
// for the ones there: its Apply reads only the item's controller (and
// its own constant fields), and passes the source, if at all, only as
// the attribution of the event it causes (CR 119.9's "a source causes
// [a player] to gain life"). A primitive that reads the source's
// characteristics or counters, acts on "this", deals damage from it,
// reads the trigger context or the item's Params, or carries a
// continuation closure (Scry.Then) is not source-blind.

// doStepsProbe is the item the registry hands a Do closure to ask which
// primitives it sequences. It is never a real item: the closure answers
// with doStepsReport before it touches the game, which is nil here.
var doStepsProbe = &game.StackItem{}

// doStepsReport is the probe's answer, carried as an error so Do's
// signature is unchanged.
type doStepsReport struct{ steps []Applier }

func (doStepsReport) Error() string { return "effects: Do steps probe" }

func doSteps(steps []Applier) error { return doStepsReport{steps: steps} }

// doClosurePC is the code pointer every Do closure shares. Do is marked
// noinline so an inlined copy of its closure cannot have a different
// one; if it ever did, doStepsOf would just say "not a Do" and the row
// would keep asking.
var doClosurePC = reflect.ValueOf(Do()).Pointer()

// doStepsOf returns the primitives e sequences when e is a Do closure.
// It calls e only after checking that it is one, and then only with the
// probe, which Do answers without touching the game.
func doStepsOf(e Effect) ([]Applier, bool) {
	if e == nil || reflect.ValueOf(e).Pointer() != doClosurePC {
		return nil, false
	}
	var rep doStepsReport
	if err := e(nil, doStepsProbe); errors.As(err, &rep) {
		return rep.steps, true
	}
	return nil, false
}

// sourceBlindStep reports whether one primitive is on the source-blind
// list. Each reads the item's controller only (a zero Player or
// Controller field means "the controller"; a set one is a constant),
// and its other fields are constants:
//
//   - GainLife: the controller gains a fixed amount. The source is
//     passed on as the gain's attribution (CR 119.9) and nothing else.
//   - DrawCards, MillCards: a fixed number of cards.
//   - GetEnergy: a fixed number of energy counters.
//   - CreateToken: a fixed number of a fixed token.
//   - BecomeTheMonarch: the controller becomes the monarch.
//   - Scry, Surveil with no Then: a fixed number of cards, the source
//     passed on as the prompt's and the event's attribution. A Then is
//     a closure the registry cannot read, so it is not on the list.
func sourceBlindStep(a Applier) bool {
	switch s := a.(type) {
	case GainLife, DrawCards, MillCards, GetEnergy, CreateToken, BecomeTheMonarch:
		return true
	case Scry:
		return s.Then == nil
	case Surveil:
		return s.Then == nil
	}
	return false
}

// sourceBlindRow reports whether a triggered row is source-blind: no
// Build (which could fill in Params, a payload or a label per source),
// no target or mode clause, and an Effect that is a non-empty Do of
// source-blind primitives.
func sourceBlindRow(t game.TriggeredAbility) bool {
	if t.Build != nil || t.Targets != nil || t.TargetsFrom != nil || t.Modes != nil || t.TargetsFromReadsBoard {
		return false
	}
	steps, ok := doStepsOf(t.Effect)
	if !ok || len(steps) == 0 {
		return false
	}
	for _, s := range steps {
		if !sourceBlindStep(s) {
			return false
		}
	}
	return true
}

// classifySourceBlind sets SourceBlind on every triggered row of a
// definition, overwriting whatever the row carried: the flag is the
// registry's to set, never a card file's.
func classifySourceBlind(d *game.CardDef) {
	if d == nil {
		return
	}
	for i := range d.Triggered {
		d.Triggered[i].SourceBlind = sourceBlindRow(d.Triggered[i])
	}
}
