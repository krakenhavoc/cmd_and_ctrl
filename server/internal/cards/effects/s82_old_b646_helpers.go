package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b646_helpers.go — Invasion, "Activated abilities" (S82 OLD batch
// 646, #5616). Append-only.

// S82OldB646TargetGainsRow is "<cost>: Target <what> gains <keyword> until end
// of turn" where what is the printed target phrase after "Target" ("black or
// green creature"). The keyword may be a protection token. preds narrow the
// creature target.
func S82OldB646TargetGainsRow(cost game.AbilityCost, costText, what, keywordText, keyword string, preds ...CardPredicate) ActivatedAbility {
	label := costText + ": Target " + what + " gains " + keywordText + " until end of turn."
	return ActivatedAbility{
		Label:   label,
		Cost:    cost,
		Targets: TargetCreature("target "+what, preds...),
		Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, keyword)),
		Effect:  EbTargetCreatureGainsUntilEOT(keyword, label),
	}
}

// S82OldB646TargetBecomesBlue is "Target creature becomes blue until end of
// turn" (Metathran Transport): a layer 5 colour change that replaces the
// creature's colours (CR 613.1e).
func S82OldB646TargetBecomesBlue(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return UntilEndOfTurn(ctx, id, nil, "Metathran Transport — blue until end of turn", game.SetColorsMod(ctx.TextColors("U")...))
}

// s82OldB646ThisColorChosen is the colour answer of "this creature becomes the
// color of your choice until end of turn" (Rainbow Crow, Kavu Chameleon): the
// tail runs against the asking item and does nothing once the source is a new
// object (CR 400.7). The label's card name is the tail's one Word after the
// colour.
var s82OldB646ThisColorChosen = CardTail("s82-old-b646/this-color-chosen", func(ctx *Context, o game.TailOutcome) error {
	color := o.Tail.Word
	if color == "" || len(o.Tail.Words) == 0 || !SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return ScopedEffectFor{
		Target:   ctx.Item.SourceCardID,
		Mods:     []game.Mod{game.SetColorsMod(color)},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    o.Tail.Words[0] + " — " + game.ColorName(color),
	}.Apply(ctx)
})

// S82OldB646ThisBecomesChosenColor is "This creature becomes the color of your
// choice until end of turn."
func S82OldB646ThisBecomesChosenColor(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		ctx := NewContext(g, item)
		then := ctx.ItemTail(s82OldB646ThisColorChosen)
		then.Words = []string{name}
		ChooseColorTail(game.ColorForBenefit, g, item.Controller, item.SourceCardID, name+" — choose a color", then)
		return nil
	}
}

// s82OldB646TargetColorChosen is the colour answer of "Target creature
// becomes the color of your choice until end of turn" (Tidal Visionary): the
// creature is the tail's one carried ID.
var s82OldB646TargetColorChosen = CardTail("s82-old-b646/target-color-chosen", func(ctx *Context, o game.TailOutcome) error {
	color := o.Tail.Word
	if color == "" || len(o.Tail.Carry) == 0 || len(o.Tail.Words) == 0 {
		return nil
	}
	return ScopedEffectFor{
		Target:   o.Tail.Carry[0],
		Mods:     []game.Mod{game.SetColorsMod(color)},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    o.Tail.Words[0] + " — " + game.ColorName(color),
	}.Apply(ctx)
})

// S82OldB646TargetBecomesChosenColor is "Target creature becomes the color of
// your choice until end of turn": the target is chosen on activation, the
// colour as the ability resolves.
func S82OldB646TargetBecomesChosenColor(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		then := ctx.ItemTail(s82OldB646TargetColorChosen)
		then.Carry = []uuid.UUID{id}
		then.Words = []string{name}
		ChooseColorTail(game.ColorForBenefit, g, item.Controller, item.SourceCardID, name+" — choose a color", then)
		return nil
	}
}

// S82OldB646SacrificeACreature is "sacrifice a creature" with no "if you
// can't" rider: the controller picks any creature they control, this one
// included (Devouring Strossus's upkeep trigger).
func S82OldB646SacrificeACreature(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller,
			SacrificeSpec("a creature", Creature()), name+" — sacrifice a creature")
		return nil
	}
}

// S82OldB646ExileThisAndTarget is "Exile this creature and target creature
// …" (Hunting Kavu): each object that is still there is exiled; the target is
// re-checked as the ability resolves (CR 608.2b) and the source only while it
// is still the permanent that activated (CR 400.7).
func S82OldB646ExileThisAndTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if SourceIsStillThisPermanent(g, item) {
		if err := (ExileTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
			return err
		}
	}
	if id, ok := B16FirstLegalTargetCard(ctx); ok {
		return ExileTarget{Target: id}.Apply(ctx)
	}
	return nil
}
