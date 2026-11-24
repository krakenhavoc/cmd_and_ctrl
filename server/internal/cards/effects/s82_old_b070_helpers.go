package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b070_helpers.go — shared bodies for Kaldheim batch 070.

// ForetellFromYourHand is Dream Devourer's first sentence, one grant:
//
//	"Each nonland card in your hand without foretell has foretell. Its
//	 foretell cost is equal to its mana cost reduced by {2}."
//
//	SpecialActionGrants: []game.SpecialActionGrant{ForetellFromYourHand(2)},
func ForetellFromYourHand(reducedBy int) game.SpecialActionGrant {
	return game.SpecialActionGrant{
		Kind:           game.SpecialActionForetell,
		Zone:           game.ZoneHand,
		Nonland:        true,
		CostIsManaCost: true,
		CostReducedBy:  reducedBy,
	}
}

// YouForetoldACard is "whenever you foretell a card" (CR 702.143a): the
// special action the permanent's controller took is a foretell, which
// PerformSpecialAction logs with the keyword's label. The trigger is the
// action, not the card landing in exile.
func YouForetoldACard(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventSpecialAction && ev.Actor == source.Controller && ev.Label == "Foretell "+game.ForetellExileCost
}

// OldB070CreatureOrEnchantmentSpellMana is Open the Omenpaths' spend
// restriction: "Spend this mana only to cast creature or enchantment spells."
func OldB070CreatureOrEnchantmentSpellMana() []string {
	return []string{ManaRestrictCast, ManaRestrictAnyType("Creature", "Enchantment")}
}

// OldB070OmenpathsFirst adds the two mana of the first colour and asks for the
// second, which may not be the first. An on-disk key: never renamed, never
// reused.
var OldB070OmenpathsFirst = CardTail("tail/old-b070-open-the-omenpaths-first-color", func(ctx *Context, o game.TailOutcome) error {
	if err := (AddMana{Player: o.Tail.Player, Produced: "{" + o.Tail.Word + "}{" + o.Tail.Word + "}",
		Restrictions: OldB070CreatureOrEnchantmentSpellMana()}).Apply(ctx); err != nil {
		return err
	}
	ChooseColorOtherThanTail(game.ColorForMana, ctx.Game, o.Tail.Player, o.Tail.Source,
		"Open the Omenpaths — choose the other color for two more mana", o.Tail.Word,
		game.KeyedTail{Ref: OldB070OmenpathsSecond, Player: o.Tail.Player, Source: o.Tail.Source})
	return nil
})

// OldB070OmenpathsSecond adds the two mana of the second colour. An on-disk key.
var OldB070OmenpathsSecond = CardTail("tail/old-b070-open-the-omenpaths-second-color", func(ctx *Context, o game.TailOutcome) error {
	return AddMana{Player: o.Tail.Player, Produced: "{" + o.Tail.Word + "}{" + o.Tail.Word + "}",
		Restrictions: OldB070CreatureOrEnchantmentSpellMana()}.Apply(ctx)
})
