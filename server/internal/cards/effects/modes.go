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

// YouControlACommander is the Commander Legends Will cycle's
// conditional mode count — "If you control a commander as you cast
// this spell, you may choose both instead":
// ChooseOne(…).OrUpToIf(2, YouControlACommander) (#1590). It is the
// free-spell cycle's controlsACommander registered by name, so "control
// a commander" reads the same in both places: a commander PERMANENT
// the chooser controls, anybody's.
var YouControlACommander = game.ModeCondition("you-control-a-commander", controlsACommander)

// YouControlAWizard is Flame of Anor's — "If you control a Wizard as
// you cast this spell, you may choose two instead". ControlsA reads
// effective subtypes, so a changeling counts.
var YouControlAWizard = game.ModeCondition("you-control-a-wizard", ControlsA("Wizard"))

// WasKicked is the kicker cards' count — "If this spell was kicked,
// choose any number instead" (the Inscription cycle), "If it was
// kicked, choose both instead" (Depth Defiler). It reads the kicker
// announced WITH the modes (CR 601.2b), not the board, which is why it
// is registered through ModeConditionOnAnnouncement (#1655). For a
// "when you cast this spell" trigger the engine hands it the spell's
// own record, so the same condition serves both.
var WasKicked = game.ModeConditionOnAnnouncement("was-kicked", func(_ *game.Game, q game.ModeCountQuery) bool {
	return q.Kicked()
})

// DeliriumForModes is "If there are four or more card types among
// cards in your graveyard, choose both instead" (Prophetic Titan) —
// delirium, read for the chooser as the modes are chosen (#1655).
var DeliriumForModes = game.ModeCondition("delirium", func(g *game.Game, chooser uuid.UUID) bool {
	return b16CardTypesInGraveyard(g, chooser) >= 4
})

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

// ChooseOneNotChosenThisTurn — "Choose one that hasn't been chosen
// this turn —" (ADR 0097: Gala Greeters, Monument to Endurance,
// Kargan Intimidator). For a TRIGGERED or ACTIVATED ability only:
// each bullet may be chosen once per turn by this object's ability,
// and an instance with none left is removed (CR 700.2b). The engine
// keeps the memory — per object, recorded when the mode is chosen,
// kept across a change of control — so the card file declares the
// restriction and nothing else.
func ChooseOneNotChosenThisTurn(options ...game.ModeOption) *game.ModeSpec {
	return &game.ModeSpec{
		Prompt: "Choose one that hasn't been chosen this turn", Options: options, Min: 1, Max: 1,
		NotChosen: game.ModeMemoryThisTurn,
	}
}

// ChooseOneNotChosen — "Choose one that hasn't been chosen —" with no
// duration (ADR 0097: Silent Hallcreeper, Demonic Pact): each bullet
// once, ever, for this object's ability. The memory ends when the
// object does (CR 400.7), so a permanent that leaves and returns may
// choose every bullet again.
func ChooseOneNotChosen(options ...game.ModeOption) *game.ModeSpec {
	return &game.ModeSpec{
		Prompt: "Choose one that hasn't been chosen", Options: options, Min: 1, Max: 1,
		NotChosen: game.ModeMemoryEver,
	}
}

// SpreeMode declares one Spree bullet (CR 702.172a): its own printed
// additional cost — mana, brace notation — paid only if this bullet
// is chosen, on top of the spell's own cost and every other chosen
// bullet's. Its OnResolve branches on ctx.HasMode(i), the older
// modal-spell shape; use SpreeModeDoing when the bullet reads better
// as its own closure.
func SpreeMode(label, cost string, targets ...*game.TargetSpec) game.ModeOption {
	o := Mode(label, targets...)
	o.Cost = cost
	return o
}

// SpreeModeDoing is SpreeMode plus ModeDoing's own body — Three Steps
// Ahead's three unrelated bullets are easier to read as three
// closures than as one long if-chain in OnResolve.
func SpreeModeDoing(label, cost string, targets *game.TargetSpec, effect func(item *game.StackItem, ctx *Context, occurrence int) error) game.ModeOption {
	o := ModeDoing(label, targets, effect)
	o.Cost = cost
	return o
}

// Spree — CR 702.172a: "Choose one or more modes. As an additional
// cost to cast this spell, pay the costs associated with those modes
// chosen this way." Not repeatable: no printed Spree card allows
// choosing the same bullet twice.
func Spree(options ...game.ModeOption) *game.ModeSpec {
	return &game.ModeSpec{Prompt: "Spree (choose one or more)", Options: options, Min: 1, Max: len(options)}
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

// OptionTargets is every still-legal target (CR 608.2b) announced for
// the occurrence(s) that chose option `option`, for the older
// OnResolve shape — a run of `if ctx.HasMode(i)` blocks in PRINTED
// order (CR 608.2c) — once a card may take more than one bullet and
// each bullet has its own target group (#764). Reading item.Targets[0]
// instead is right only while exactly one bullet can be chosen: with
// both chosen it hands the second bullet the first one's target.
// Added for the conditional-mode-count Wills (#1590).
func OptionTargets(ctx *Context, option int) []game.TargetRef {
	var out []game.TargetRef
	for occ, m := range ctx.Modes() {
		if m != option {
			continue
		}
		for _, t := range ctx.ModeTargets(occ) {
			if ctx.IsTargetLegal(t) {
				out = append(out, t)
			}
		}
	}
	return out
}

// BulletsInPrintedOrder runs each chosen bullet's body in PRINTED
// order (CR 608.2c) — option 0's occurrences first, then option 1's —
// whatever order the caster clicked them in. `bodies[i]` is option i's
// body, handed the occurrence so it reads its own target group.
//
// For a card whose bullets can see each other (#1655: Depth Defiler's
// bounce before its draw-then-discard, Inscription of Abundance's
// counters before its "greatest power"), where the engine's ModeOption
// Effect walk — announce order — would let the caster reorder them.
// Declare the bullets with Mode, not ModeDoing, and call this from
// OnResolve (a spell) or the ability's Effect (a trigger).
func BulletsInPrintedOrder(item *game.StackItem, ctx *Context, bodies ...func(item *game.StackItem, ctx *Context, occ int) error) error {
	modes := ctx.Modes()
	for opt, body := range bodies {
		for occ, m := range modes {
			if m != opt || body == nil {
				continue
			}
			if err := body(item, ctx, occ); err != nil {
				return err
			}
		}
	}
	return nil
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

// TokenCopyTheModesTarget is "create a token that's a copy of target
// <thing> you control" as a modal bullet's body — Sublime Epiphany's
// creature bullet and Three Steps Ahead's artifact-or-creature bullet
// differ only in the clause on the option.
func TokenCopyTheModesTarget(item *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return CreateTokenCopy{Controller: item.Controller, Copy: t.ID, N: 1}.Apply(ctx)
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
// StackItem.Kind (#1211). Three Steps Ahead's Spree bullet (S45)
// reuses it too, differing only in the mode's own Cost.
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
