package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// remains_tapped.go — the "for as long as this remains tapped" family
// (ADR 0109 §3, #1894, and owner decision 4):
//
//	"{2}, {T}: Target creature gets +0/+3 for as long as this artifact
//	 remains tapped."                                          Endoskeleton
//	"{2}, {T}: Gain control of target creature with power less than or
//	 equal to the number of Islands you control for as long as this
//	 artifact remains tapped."                            Vedalken Shackles
//	"{T}: Gain control of target creature whose controller controls an
//	 Island for as long as you control this creature and this creature
//	 remains tapped."                                             Seasinger
//
// Every one of these is a continuous effect from a resolving ability
// (CR 611.2), so it is a ScopedEffect record — data a restore point
// carries — timed by an ADR 0063 duration. The second form is the
// conjunction ADR 0109 §3 adds (Duration.Also): the effect ends the
// moment EITHER half stops (CR 611.2b). Each card also prints "You may
// choose not to untap ~ during your untap step", the untap opt-out
// (mayChooseNotToUntapSelf), which is how its controller keeps it.
//
// "Never starts" (CR 611.2b): a source that has untapped, left, or come
// back as a new object (CR 400.7, #1432) before the ability resolves
// makes the duration false as it would begin, so nothing is registered.

// DurationWhileThisRemainsTapped is "for as long as ~ remains tapped",
// about the resolving ability's source. False when the duration would
// never start.
func DurationWhileThisRemainsTapped(ctx *Context) (game.Duration, bool) {
	if ctx.isNewSourceObjectAsThis(ctx.Source()) { // #1432
		return game.Duration{}, false
	}
	return ctx.Game.ForAsLongAsSourceTappedDuration(ctx.Source())
}

// WhileYouControlThisAndItRemainsTapped is "for as long as you control
// ~ and ~ remains tapped" (ADR 0109 §3): "you" is the ability's
// controller. False when either half is already false, so the card
// registers nothing (CR 611.2b).
func WhileYouControlThisAndItRemainsTapped(ctx *Context) (game.Duration, bool) {
	if ctx.isNewSourceObjectAsThis(ctx.Source()) { // #1432
		return game.Duration{}, false
	}
	return ctx.Game.ForAsLongAsYouControlAndSourceTappedDuration(ctx.Source(), ctx.Controller())
}

// TargetGetsWhileThisRemainsTapped is the Effect of "target <permanent>
// gets <mods> for as long as this remains tapped": the Couriers'
// "+2/+2 and has <keyword>", Endoskeleton's "+0/+3", Hisoka's Guard's
// "has shroud". The ability's Targets clause says what may be chosen;
// a target gone by resolution changes nothing (CR 608.2b).
func TargetGetsWhileThisRemainsTapped(label string, mods ...game.Mod) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		d, ok := DurationWhileThisRemainsTapped(ctx)
		if !ok {
			return nil
		}
		return ScopedEffectFor{Target: target, Mods: mods, Duration: d, Label: label}.Apply(ctx)
	}
}

// TapTargetsAndHoldWhileThisRemainsTapped is the Effect of "tap target
// <permanent>. It doesn't untap during its controller's untap step for
// as long as this remains tapped" (Mana Leech, Sand Squid, Ice Floe):
// Rust Tick's hold over the ability's still-legal targets.
func TapTargetsAndHoldWhileThisRemainsTapped(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return TapAndHoldWhileThisRemainsTapped(ctx, holdTargetIDs(ctx))
}

// GainControlOfTargetFor is the Effect of "gain control of target
// <permanent> for as long as <duration>", with the duration built by
// `dur` as the ability resolves: DurationWhileThisRemainsTapped
// (Vedalken Shackles) or WhileYouControlThisAndItRemainsTapped
// (Seasinger). The ability's controller gains control. A duration that
// would never start takes nothing (CR 611.2b).
func GainControlOfTargetFor(label string, dur func(ctx *Context) (game.Duration, bool)) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		d, ok := dur(ctx)
		if !ok {
			return nil
		}
		return GainControl{Target: target, Controller: ctx.Controller(), Duration: d, Label: label}.Apply(ctx)
	}
}

// AttackingYou is "creature … that's attacking you" (Ice Floe): a
// creature attacking the caster as a player. One attacking a
// planeswalker or battle the caster controls is attacking that
// permanent, not "you" (CR 506.3).
func AttackingYou() CardPredicate {
	return func(_ *game.Game, caster uuid.UUID, c game.Card) bool {
		return c.AttackingTarget != uuid.Nil && c.AttackingTarget == caster
	}
}

// PowerAtMostIslandsYouControl is "creature with power less than or
// equal to the number of Islands you control" (Vedalken Shackles),
// counted as the target is chosen and again as the ability resolves
// (CR 608.2b).
func PowerAtMostIslandsYouControl() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		islands := 0
		for _, p := range g.BattlefieldCardsForEffect() {
			if p.Controller == caster && p.HasSubtype("Island") {
				islands++
			}
		}
		return c.CurrentPower() <= islands
	}
}
