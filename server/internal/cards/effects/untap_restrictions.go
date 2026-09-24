package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func doesntUntapDuringYourUntapStep() game.UntapStepRestriction {
	return game.UntapStepRestriction{
		Label: "doesn't untap during your untap step",
		Restricts: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target != nil && source != nil && target.InstanceID == source.InstanceID
		},
	}
}

func enchantedDoesntUntap() game.UntapStepRestriction {
	return game.UntapStepRestriction{Label: "enchanted creature doesn't untap", Restricts: AttachedToSource}
}

func doesntUntapDuringTheirControllersUntapSteps(match CardPredicate) game.UntapStepRestriction {
	return game.UntapStepRestriction{
		Label: "doesn't untap during its controller's untap step",
		Restricts: func(target *game.Card, g *game.Game, _ *game.Card) bool {
			return target != nil && (match == nil || match(g, target.Controller, *target))
		},
	}
}

// DoesntUntapNextUntapStep records the fixed set from a resolved one-shot
// effect. Player uuid.Nil means each permanent's current controller.
type DoesntUntapNextUntapStep struct {
	Targets []uuid.UUID
	Player  uuid.UUID
	Label   string
}

func (d DoesntUntapNextUntapStep) Apply(ctx *Context) error {
	for _, id := range ctx.withoutNewSourceObject(d.Targets) { // #1432
		if err := ctx.Game.SkipNextUntapForEffect(id, d.Player); err != nil {
			return err
		}
	}
	return nil
}

// TapAndFreeze implements the common "tap ...; those permanents don't
// untap during ... next untap step" resolution.
type TapAndFreeze struct {
	Targets []uuid.UUID
	Player  uuid.UUID
	Label   string
}

func (t TapAndFreeze) Apply(ctx *Context) error {
	targets := ctx.withoutNewSourceObject(t.Targets) // #1432
	for _, id := range targets {
		if err := ctx.Game.TapTargetForEffect(id); err != nil {
			return err
		}
	}
	return (DoesntUntapNextUntapStep{Targets: targets, Player: t.Player, Label: t.Label}).Apply(ctx)
}

// DoesntUntapWhile records "it doesn't untap during its controller's
// untap step for as long as <Duration>" on each target (#1313, CR
// 611.2b; ADR 0058's 2026-09-23 amendment). The hold is read at every
// untap step of the permanent's controller and ends when the duration
// does. It is never used up by a step.
//
// Build Duration with the Duration* helper whose second return answers
// CR 611.2b's "never starts", and apply nothing when it is false. The
// TapAndHold… helpers below do that for you.
type DoesntUntapWhile struct {
	Targets  []uuid.UUID
	Duration game.Duration
}

func (d DoesntUntapWhile) Apply(ctx *Context) error {
	for _, id := range ctx.withoutNewSourceObject(d.Targets) { // #1432
		if err := ctx.Game.HoldUntappedForEffect(id, d.Duration); err != nil {
			return err
		}
	}
	return nil
}

// TapAndHoldWhileYouControlThis is "tap target creature. It doesn't
// untap during its controller's untap step for as long as you control
// ~" — Ty Lee, Chi Blocker; Dungeon Geists; Tidebinder Mage. "You" is
// the ability's controller and "~" its source.
//
// The tap always happens. The hold is recorded only if the duration
// starts: if the source has already left the battlefield or changed
// controller, the target is tapped and untaps as normal (CR 611.2b,
// and the Dungeon Geists ruling of 2019-07-12).
func TapAndHoldWhileYouControlThis(ctx *Context, targets []uuid.UUID) error {
	if err := tapEach(ctx, targets); err != nil {
		return err
	}
	d, ok := DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller())
	if !ok {
		return nil
	}
	return DoesntUntapWhile{Targets: targets, Duration: d}.Apply(ctx)
}

// TapAndHoldWhileThisRemainsTapped is "tap target artifact. It doesn't
// untap during its controller's untap step for as long as ~ remains
// tapped" — Rust Tick, Amber Prison. Same "never starts" rule: a
// source untapped in response taps the target and holds nothing.
func TapAndHoldWhileThisRemainsTapped(ctx *Context, targets []uuid.UUID) error {
	if err := tapEach(ctx, targets); err != nil {
		return err
	}
	d, ok := ctx.Game.ForAsLongAsSourceTappedDuration(ctx.Source())
	if !ok || ctx.isNewSourceObject(ctx.Source()) { // #1432
		return nil
	}
	return DoesntUntapWhile{Targets: targets, Duration: d}.Apply(ctx)
}

func tapEach(ctx *Context, targets []uuid.UUID) error {
	for _, id := range targets {
		if err := ctx.Game.TapTargetForEffect(id); err != nil {
			return err
		}
	}
	return nil
}

// holdTargetIDs is the instance IDs of the item's still-legal targets
// (CR 608.2b), in slot order.
func holdTargetIDs(ctx *Context) []uuid.UUID {
	ts := ctx.LegalTargets()
	ids := make([]uuid.UUID, 0, len(ts))
	for _, t := range ts {
		ids = append(ids, t.ID)
	}
	return ids
}
