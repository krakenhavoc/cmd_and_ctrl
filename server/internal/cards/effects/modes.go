package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// modes.go — S20 sub-PR 4: constructors for Spec.Modes so a modal
// card file reads like its oracle text:
//
//	Modes: ChooseOne(
//		Mode("Exile target player's graveyard.", TargetPlayer("target player")),
//		Mode("Destroy target artifact.", TargetPermanent("target artifact", Artifact())),
//		Mode("Each creature deals 1 damage to its controller."),
//	),
//
// and its OnResolve is a run of `if ctx.HasMode(i)` blocks in
// printed order. The engine validates the choice at announce and
// routes the chosen option's TargetSpec through the S20 legality
// gate; the client shows the picker between the X prompt and
// targeting.

// Mode declares one option. Pass a TargetSpec when the bullet
// targets; omit it otherwise. A bullet with two differently
// constrained targets passes one statement built with Clauses.
func Mode(label string, targets ...*game.TargetSpec) game.ModeOption {
	o := game.ModeOption{Label: label}
	if len(targets) > 0 {
		o.Targets = targets[0]
	}
	return o
}

// ModeDoing declares one option together with its BODY, for a modal
// trigger or activated ability — which has no OnResolve to branch
// in. `occurrence` is the index into the announced modes, so the
// bullet reads its own target group with ctx.ModeTargets(occ)
// (#764).
//
// A modal SPELL may use this too, and then its OnResolve is nil;
// the older `if ctx.HasMode(i)` shape still works and reads the same
// data.
func ModeDoing(label string, targets *game.TargetSpec, effect func(item *game.StackItem, ctx *Context, occurrence int) error) game.ModeOption {
	return game.ModeOption{
		Label:   label,
		Targets: targets,
		Effect: func(g *game.Game, item *game.StackItem, occurrence int) error {
			return effect(item, NewContext(g, item), occurrence)
		},
	}
}

// ChooseOne — "Choose one —".
func ChooseOne(options ...game.ModeOption) *game.ModeSpec {
	return &game.ModeSpec{Prompt: "Choose one", Options: options, Min: 1, Max: 1}
}

// ChooseN — "Choose two —" (n, n) or "choose one or both" (1, 2).
func ChooseN(prompt string, min, max int, options ...game.ModeOption) *game.ModeSpec {
	return &game.ModeSpec{Prompt: prompt, Options: options, Min: min, Max: max}
}

// ChooseOneOrMore — "Choose one or more —" (Sublime Epiphany): at
// least one bullet, at most all of them, each at most once.
func ChooseOneOrMore(options ...game.ModeOption) *game.ModeSpec {
	return &game.ModeSpec{Prompt: "Choose one or more", Options: options, Min: 1, Max: len(options)}
}

// ChooseNRepeating — CR 700.2d, "you may choose the same mode more
// than once" (Mystic Confluence's "Choose three. You may choose the
// same mode more than once"). Each occurrence gets its own targets.
func ChooseNRepeating(prompt string, min, max int, options ...game.ModeOption) *game.ModeSpec {
	return &game.ModeSpec{Prompt: prompt, Options: options, Min: min, Max: max, Repeatable: true}
}

// ManaValueGE passes when the card's mana value is ≥ n (Austere
// Command's "mana value 4 or greater").
// Same reading as ManaValueLE: game.(*Game).ManaValueForEffect.
func ManaValueGE(n int) CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		mv, ok := g.ManaValueForEffect(c)
		return ok && mv >= n
	}
}

// ModeTarget is the first still-legal target announced for this mode
// OCCURRENCE (CR 608.2b), or (zero, false) when the bullet had none
// or the one it had has gone. The read every targeted bullet wants:
// a mode's targets are its own group, and a bullet must never reach
// for the group of the bullet chosen beside it.
func ModeTarget(ctx *Context, occurrence int) (game.TargetRef, bool) {
	for _, t := range ctx.ModeTargets(occurrence) {
		if ctx.IsTargetLegal(t) {
			return t, true
		}
	}
	return game.TargetRef{}, false
}

// DestroyTheModesTarget is "destroy target <thing>" as a modal
// bullet's body — shared by Kolaghan's Command's artifact bullet and
// Glissa Sunslayer's enchantment bullet, which differ only in the
// clause on the option.
func DestroyTheModesTarget(item *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return DestroyTarget{Target: t.ID}.Apply(ctx)
}

// BounceTheModesTarget is "return target <thing> to its owner's
// hand" as a modal bullet's body — Mystic Confluence's and Sublime
// Epiphany's, likewise differing only in the clause.
func BounceTheModesTarget(item *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return BounceToHand{Target: t.ID}.Apply(ctx)
}

// CounterTheModesTarget is "counter target <thing on the stack>" as a
// modal bullet's body — Sublime Epiphany's first TWO bullets, which
// differ only in their clause ("target spell" and "target activated
// or triggered ability") and not at all in what they do, because
// CounterTarget takes a stack ITEM id and discriminates on
// StackItem.Kind (#1211).
func CounterTheModesTarget(item *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return CounterTarget{StackID: t.ID}.Apply(ctx)
}

// DealFixedDamageToModesTarget is "<source> deals `amount` damage to
// <this mode's target>" as a modal bullet's body, for a printed fixed
// amount — Kolaghan's Command's and Prismari Command's "deals 2
// damage to any target" bullets share this shape (#1112).
func DealFixedDamageToModesTarget(amount int) func(item *game.StackItem, ctx *Context, occ int) error {
	return func(item *game.StackItem, ctx *Context, occ int) error {
		t, ok := ModeTarget(ctx, occ)
		if !ok {
			return nil
		}
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: amount}.Apply(ctx)
	}
}

// DealXDamageToModesTarget is "this spell deals X damage to <this
// mode's target>" as a modal bullet's body — Mishra's Command's
// creature and planeswalker bullets differ only in the target
// clause, not the body (#1112).
func DealXDamageToModesTarget(item *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: ctx.X()}.Apply(ctx)
}
