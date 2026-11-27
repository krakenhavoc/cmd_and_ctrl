package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b647_helpers.go — Invasion, "Activated abilities" (S82 OLD batch
// 647, #5616). Append-only.

// S82OldB647TargetGainsRow is "<cost>: Target creature gains <keyword> until
// end of turn". The keyword may be a protection token; keywordText is how the
// card prints it.
func S82OldB647TargetGainsRow(cost game.AbilityCost, costText, keywordText, keyword string) ActivatedAbility {
	label := costText + ": Target creature gains " + keywordText + " until end of turn."
	return ActivatedAbility{
		Label:   label,
		Cost:    cost,
		Targets: TargetCreature("target creature"),
		Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, keyword)),
		Effect:  EbTargetCreatureGainsUntilEOT(keyword, label),
	}
}

// S82OldB647TapTargetRow is "<cost>: Tap target creature." with a {T} in the
// cost (Benalish Trapper, Thornscape Apprentice).
func S82OldB647TapTargetRow(mana, costText string) ActivatedAbility {
	return ActivatedAbility{
		Label:   costText + ": Tap target creature.",
		Cost:    Plus(ManaCost(mana), TapCost()),
		Targets: TargetCreature("target creature"),
		Effect:  TapChosenPermanent,
	}
}

// S82OldB647ExchangeThisWithTarget is "Exchange control of this creature and
// target creature" (Phyrexian Infiltrator): the exchange needs both objects
// still where they were (CR 701.12a), which ExchangeControl reports and does
// nothing without.
func S82OldB647ExchangeThisWithTarget(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return ExchangeControl{A: item.SourceCardID, B: id, Label: name + " — exchange control"}.Apply(ctx)
	}
}

// S82OldB647LookAtHandAndDiscard is "Look at target player's hand and choose
// a card from it. That player discards that card" (Seer's Vision): only the
// chooser looks.
func S82OldB647LookAtHandAndDiscard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return ChooseFromRevealedHand{Player: TargetedPlayer(ctx), Label: "card", Look: true}.Apply(ctx)
}

// s82OldB647BlindSeerColorChosen sets the chosen colour on the target spell or
// permanent. An on-disk key.
var s82OldB647BlindSeerColorChosen = CardTail("s82-old-b647/blind-seer-color-chosen", func(ctx *Context, o game.TailOutcome) error {
	color := o.Tail.Word
	if color == "" {
		return nil
	}
	return Gap2B05TargetBecomesColor(ctx, []string{color}, "Blind Seer — "+game.ColorName(color)+" until end of turn")
})

// S82OldB647TargetSpellOrPermanentBecomesChosenColor is "Target spell or
// permanent becomes the color of your choice until end of turn" (Blind Seer):
// the colour is chosen as the ability resolves, and only if the target is
// still legal (CR 608.2b).
func S82OldB647TargetSpellOrPermanentBecomesChosenColor(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(ctx.LegalTargets()) == 0 {
		return nil
	}
	ChooseColorTail(game.ColorForBenefit, g, item.Controller, item.SourceCardID,
		"Blind Seer — choose a color", ctx.ItemTail(s82OldB647BlindSeerColorChosen))
	return nil
}
