package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// footprint.go — what a triggered row's effect does, as data the
// trigger-order independence check can read (#2884, ADR 0018's #2884
// amendment, game/trigger_independence.go).
//
// The registry reads the same thing source_blind.go reads: a row whose
// Effect is an effects.Do answers a probe with its steps. Each step that
// can say what it touches implements footprinter, and a row's
// game.TriggeredAbility.Footprint is the list of those answers. One step
// that cannot answer, or a row with a "you may", a mode clause
// or a clause built from the trigger, leaves the footprint nil, and the
// row's triggers keep their prompt. A row with a Build fill-in still
// gets the footprint of its declared Effect; the check accepts an item
// it built only when the item carries nothing the footprint does not
// read (game.builtItemIsPlain).
//
// Adding a primitive here needs the argument the amendment makes for
// the ones there: every object and player it touches is the item's
// controller, the item's source or the item's chosen target, its other
// fields are constants, and the game kind it declares says exactly what
// it reads and writes.

// footprinter is a primitive that can declare its step.
type footprinter interface {
	footprint() (game.FootprintStep, bool)
}

func (s GainLife) footprint() (game.FootprintStep, bool) {
	return game.FootprintStep{Kind: game.FootprintGainLife, N: s.Amount}, s.Player == uuid.Nil && s.Amount > 0
}

func (s DrawCards) footprint() (game.FootprintStep, bool) {
	return game.FootprintStep{Kind: game.FootprintDraw, N: s.N}, s.Player == uuid.Nil && s.N > 0
}

func (s MillCards) footprint() (game.FootprintStep, bool) {
	return game.FootprintStep{Kind: game.FootprintMill, N: s.N}, s.Player == uuid.Nil && s.N > 0
}

func (s Scry) footprint() (game.FootprintStep, bool) {
	return game.FootprintStep{Kind: game.FootprintScry, N: s.N}, s.Player == uuid.Nil && s.N > 0 && s.Then == nil
}

func (s Surveil) footprint() (game.FootprintStep, bool) {
	return game.FootprintStep{Kind: game.FootprintSurveil, N: s.N}, s.Player == uuid.Nil && s.N > 0 && s.Then == nil
}

func (s GetEnergy) footprint() (game.FootprintStep, bool) {
	return game.FootprintStep{Kind: game.FootprintEnergy, N: s.N}, s.Player == uuid.Nil && s.N > 0
}

func (s CreateToken) footprint() (game.FootprintStep, bool) {
	key, manual := game.FootprintTokenOf(s.Template)
	return game.FootprintStep{Kind: game.FootprintCreateToken, N: s.N, TokenKey: key, TokenManual: manual}, s.Controller == uuid.Nil && s.N > 0
}

func (s BecomeTheMonarch) footprint() (game.FootprintStep, bool) {
	return game.FootprintStep{Kind: game.FootprintMonarch}, s.Player == uuid.Nil
}

func (s CounterOnThis) footprint() (game.FootprintStep, bool) {
	return game.FootprintStep{Kind: game.FootprintCounterOnSource, N: s.N, Counter: s.Kind}, s.N > 0 && s.Kind != ""
}

func (s DamageEachOpponent) footprint() (game.FootprintStep, bool) {
	return game.FootprintStep{Kind: game.FootprintDamageEachOpponent, N: s.N}, s.N > 0
}

func (ExileChosenTarget) footprint() (game.FootprintStep, bool) {
	return game.FootprintStep{Kind: game.FootprintExileTargets}, true
}

// footprintOfRow is a row's declared footprint, or nil when the row is
// anything the check cannot describe.
func footprintOfRow(t game.TriggeredAbility) []game.FootprintStep {
	if t.OptionalPrompt != nil || t.Modes != nil || t.TargetsFrom != nil || t.TargetsFromReadsBoard {
		return nil
	}
	steps, ok := doStepsOf(t.Effect)
	if !ok || len(steps) == 0 {
		return nil
	}
	out := make([]game.FootprintStep, 0, len(steps))
	for _, s := range steps {
		f, ok := s.(footprinter)
		if !ok {
			return nil
		}
		step, ok := f.footprint()
		if !ok {
			return nil
		}
		// A step that acts on the chosen target needs a declared
		// target clause to choose it from.
		if step.Kind == game.FootprintExileTargets && t.Targets == nil {
			return nil
		}
		out = append(out, step)
	}
	return out
}

// classifyFootprint sets Footprint on every triggered row of a
// definition, overwriting whatever the row carried: like SourceBlind,
// the field is the registry's to set, never a card file's.
func classifyFootprint(d *game.CardDef) {
	if d == nil {
		return
	}
	for i := range d.Triggered {
		d.Triggered[i].Footprint = footprintOfRow(d.Triggered[i])
	}
}

// CounterOnThis is "put N <kind> counters on this permanent" in a
// triggered ability's effect: the item's source, and only while it is
// still on the battlefield. A source that has left gets nothing (there
// is no permanent to put them on), and a source that left and came
// back is a new object that is not "this" (CR 400.7, AddCounter's own
// guard).
type CounterOnThis struct {
	Kind string
	N    int
}

func (s CounterOnThis) Apply(ctx *Context) error {
	src := ctx.Source()
	if !onBattlefield(ctx.Game, src) {
		return nil
	}
	return AddCounter{Target: src, Kind: s.Kind, N: s.N}.Apply(ctx)
}

// DamageEachOpponent is "<this> deals N damage to each opponent": one
// damage event from the item's source to every opponent of its
// controller still in the game.
type DamageEachOpponent struct {
	N int
}

func (s DamageEachOpponent) Apply(ctx *Context) error {
	return damageToEachOpponent(ctx.Game, ctx.Item, s.N)
}

// ExileChosenTarget is "exile [the] target <whatever>": the first card
// among the item's targets that is still legal as it resolves (CR
// 608.2b), or nothing.
type ExileChosenTarget struct{}

func (ExileChosenTarget) Apply(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ExileTarget{Target: t.ID}.Apply(ctx)
	}
	return nil
}
