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

// enchantedCantBecomeUntapped is "enchanted creature can't become
// untapped" (CR 614.17, #4574): Frozen in Ice. Stronger than
// enchantedDoesntUntap: an untap effect does nothing to the creature
// either, for as long as the Aura is on it.
func enchantedCantBecomeUntapped() game.UntapStepRestriction {
	return game.UntapStepRestriction{
		Label:              "enchanted creature can't become untapped",
		Restricts:          AttachedToSource,
		CantBecomeUntapped: true,
	}
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
//
// Steps is how many of those untap steps it sits out; zero is one
// ("next untap step"), two is Telekinesis (#2029).
type DoesntUntapNextUntapStep struct {
	Targets []uuid.UUID
	Player  uuid.UUID
	Label   string
	Steps   int
}

func (d DoesntUntapNextUntapStep) Apply(ctx *Context) error {
	for _, id := range ctx.withoutNewSourceObject(d.Targets) { // #1432
		if err := ctx.Game.SkipNextUntapStepsForEffect(id, d.Player, max(d.Steps, 1)); err != nil {
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

// CantBecomeUntappedWhile records "it can't become untapped for as long
// as <Duration>" on each target (#4574, CR 614.17, CR 611.2b). Unlike
// DoesntUntapWhile the permanent stays tapped through EVERY untap, not
// only its controller's untap step. Build Duration as for DoesntUntapWhile
// and apply nothing when it says false.
type CantBecomeUntappedWhile struct {
	Targets  []uuid.UUID
	Duration game.Duration
}

func (d CantBecomeUntappedWhile) Apply(ctx *Context) error {
	for _, id := range ctx.withoutNewSourceObject(d.Targets) { // #1432
		if err := ctx.Game.CantUntapForEffect(id, d.Duration); err != nil {
			return err
		}
	}
	return nil
}

// TapAndLockWhileYouControlThis is "tap target creature. It can't become
// untapped for as long as you control ~" (Spider-Woman, Secret Agent). The
// tap always happens; the lock is recorded only if the duration starts,
// exactly as TapAndHoldWhileYouControlThis does.
func TapAndLockWhileYouControlThis(ctx *Context, targets []uuid.UUID) error {
	if err := tapEach(ctx, targets); err != nil {
		return err
	}
	d, ok := DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller())
	if !ok {
		return nil
	}
	return CantBecomeUntappedWhile{Targets: targets, Duration: d}.Apply(ctx)
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
	d, ok := DurationWhileThisRemainsTapped(ctx) // #1432 included
	if !ok {
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

// PermanentsDontUntapNextStep is the player-wide "<types> <player>
// controls don't untap during <player>'s next untap step" (#3995, CR
// 502.3): Mana Vapors' lands, Misstep's creatures, Exhaustion's
// creatures and lands, the Amonkhet Last cycle's "Lands you control
// don't untap during your next untap step".
//
// Unlike DoesntUntapNextUntapStep it names no permanent. It is a rule
// read at that player's next untap step, so it covers a land played or a
// creature gained after the spell resolved (CR 611.2c: it changes no
// characteristic, so its set is not locked), and nothing a player
// stops controlling before then.
//
// Player is whose untap step; Types are lowercase permanent types, any
// of them. AnyController is Blinding Beam's bare "Creatures don't untap
// during target player's next untap step": every permanent of those
// types that would untap in that step, a Seedborn Muse's included, not
// only the ones Player controls. A Player who has left the game is
// recorded nothing.
type PermanentsDontUntapNextStep struct {
	Player        uuid.UUID
	Types         []string
	AnyController bool
	Label         string
}

// Apply implements Primitive.
func (p PermanentsDontUntapNextStep) Apply(ctx *Context) error {
	label := p.Label
	if label == "" && ctx.Item != nil {
		label = ctx.Item.Label
	}
	ctx.Game.SkipNextUntapStepForPlayerForEffect(p.Player, ctx.Source(), p.Types, p.AnyController, label)
	return nil
}

// LandsYouControlDontUntapNextStep is the Amonkhet Last cycle's "Lands
// you control don't untap during your next untap step", "you" being the
// resolving item's controller.
func LandsYouControlDontUntapNextStep(ctx *Context) error {
	return PermanentsDontUntapNextStep{
		Player: ctx.Controller(),
		Types:  []string{"land"},
		Label:  "Lands you control don't untap during your next untap step",
	}.Apply(ctx)
}
