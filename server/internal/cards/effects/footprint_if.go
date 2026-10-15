package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// footprint_if.go — an intervening "if" as a Do step (#3033, ADR 0018's
// #3033 amendment).
//
// "At the beginning of each end step, if you gained 3 or more life this
// turn, draw a card." The condition is checked as the ability triggers
// (the row's AppliesTo) and again as it resolves (CR 603.4), and the
// ability does nothing if it no longer holds. The second check is If:
//
//	Do(If{Cond: CondLifeGained{AtLeast: 3}, Then: []Applier{DrawCards{N: 1}}})
//
// Each Condition is a named value that declares the state it reads
// (game.FootprintCond). If's footprint is that read followed by the
// footprint of every Then step, declared as if it ran: the check
// conflicts the read with any other item that can change that state
// (game.addConditionReadLocked), and treats the guarded steps the way it
// treats the same steps unguarded. A Then step that cannot declare a
// footprint leaves the row without one, as in a plain Do.
//
// A Condition joins only with the argument the check makes for its
// game.FootprintCond: what it reads, and why nothing else can change its
// answer. Its Holds must read nothing more than that.

// Condition is an intervening "if" that If re-checks as the item
// resolves.
type Condition interface {
	// Holds is the condition, read as the item resolves.
	Holds(ctx *Context) bool
	// condition declares the state it reads, as a FootprintCondition
	// step, or false when it cannot.
	condition() (game.FootprintStep, bool)
}

// If is "if <Cond>, <Then>" inside a triggered ability's effect: the
// Then steps run, in order, only when Cond holds as the item resolves.
type If struct {
	Cond Condition
	Then []Applier
}

func (s If) Apply(ctx *Context) error {
	if s.Cond == nil || !s.Cond.Holds(ctx) {
		return nil
	}
	for _, step := range s.Then {
		if err := step.Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// footprints is If's footprint: the condition's read, then every Then
// step's.
func (s If) footprints() ([]game.FootprintStep, bool) {
	if s.Cond == nil || len(s.Then) == 0 {
		return nil, false
	}
	cond, ok := s.Cond.condition()
	if !ok {
		return nil, false
	}
	out := []game.FootprintStep{cond}
	for _, step := range s.Then {
		more, ok := stepFootprints(step)
		if !ok {
			return nil, false
		}
		out = append(out, more...)
	}
	return out, true
}

func condStep(c game.FootprintCond, n int) game.FootprintStep {
	return game.FootprintStep{Kind: game.FootprintCondition, Cond: c, N: n}
}

// CondLifeGained is "if you gained [N or more] life this turn":
// the item's controller's tally of life gained this turn (The Gaffer,
// Crested Sunmare). AtLeast 0 means 1.
type CondLifeGained struct{ AtLeast int }

func (c CondLifeGained) Holds(ctx *Context) bool {
	return b15LifeGainedThisTurn(ctx.Game, ctx.Controller()) >= max(c.AtLeast, 1)
}

func (c CondLifeGained) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondLifeGained, max(c.AtLeast, 1)), true
}

// CondLifeLost is "if you lost [N or more] life this turn".
// AtLeast 0 means 1.
type CondLifeLost struct{ AtLeast int }

func (c CondLifeLost) Holds(ctx *Context) bool {
	return b18LifeLostThisTurn(ctx.Game, ctx.Controller()) >= max(c.AtLeast, 1)
}

func (c CondLifeLost) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondLifeLost, max(c.AtLeast, 1)), true
}

// CondCreatureDied is "if a creature died under
// your control this turn" (Faramir, Field Commander).
type CondCreatureDied struct{}

func (CondCreatureDied) Holds(ctx *Context) bool {
	return aCreatureDiedUnderYourControlThisTurn(ctx.Game, ctx.Controller())
}

func (CondCreatureDied) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondCreatureDied, 0), true
}

// CondCitysBlessing is "if you have the city's blessing"
// as it resolves (Deadeye Brawler, Mausoleum Harpy).
type CondCitysBlessing struct{}

func (CondCitysBlessing) Holds(ctx *Context) bool {
	return YouHaveTheCitysBlessing(ctx.Game, ctx.Controller())
}

func (CondCitysBlessing) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondCitysBlessing, 0), true
}

// CondNoSpellsCast is "if you cast no spells this turn" (Edgar,
// Moonlit Sovereign).
type CondNoSpellsCast struct{}

func (CondNoSpellsCast) Holds(ctx *Context) bool {
	return ctx.Game.CastTallyFor(ctx.Controller()).Total == 0
}

func (CondNoSpellsCast) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondNoSpellsCast, 0), true
}

// CondMonarch is "if you're the monarch" (Ginger, Queen of
// Sweets).
type CondMonarch struct{}

func (CondMonarch) Holds(ctx *Context) bool {
	return YoureTheMonarch(ctx.Game, ctx.Controller())
}

func (CondMonarch) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondMonarch, 0), true
}

// CondFewerCountersOnThis is "if this permanent has fewer than N <Kind>
// counters on it" (Runaway Steam-Kin): false once the source has left
// the battlefield.
type CondFewerCountersOnThis struct {
	Kind string
	N    int
}

func (c CondFewerCountersOnThis) Holds(ctx *Context) bool {
	src := ctx.Source()
	card, ok := ctx.Game.LookupCardForEffect(src)
	return ok && onBattlefield(ctx.Game, src) && card.Counters[c.Kind] < c.N
}

func (c CondFewerCountersOnThis) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondSource, c.N), c.Kind != "" && c.N > 0
}

// CondPowerAtLeast is "if you control a creature with
// power N or greater" (Garruk's Uprising, Bugenhagen): current power,
// counters and anthems included.
type CondPowerAtLeast struct{ N int }

func (c CondPowerAtLeast) Holds(ctx *Context) bool {
	return b24ControlsCreatureWithPowerAtLeast(ctx.Game, ctx.Controller(), c.N)
}

func (c CondPowerAtLeast) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondBoard, c.N), c.N > 0
}

// CondBoard is an intervening "if" about the permanents on the
// battlefield that has no name of its own: Field of the Dead's seven land
// names, Lux Artillery's thirty counters. Test may read only the
// permanents on the battlefield (through the layers), the item's
// controller, source and triggering event, and this turn's event log; it
// declares a read of the whole battlefield (game.FootprintCondBoard).
type CondBoard struct {
	// Name says what it checks, for a reader of the card file.
	Name string
	Test func(ctx *Context) bool
}

func (c CondBoard) Holds(ctx *Context) bool { return c.Test != nil && c.Test(ctx) }

func (c CondBoard) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondBoard, 0), c.Test != nil
}
