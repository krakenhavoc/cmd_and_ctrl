package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b006_b_helpers.go — S82 old sets, batch 006, Estrid, the Masked
// (#5462).

// OldB006MaskToken is the white Aura enchantment token named Mask that
// Estrid's −1 creates: "enchant permanent" and umbra armor (the dump prints
// the same token with the older name, totem armor, CR 702.89b). Its slug is
// an on-disk identity: never renamed.
func OldB006MaskToken() game.Card { return TokenFromCatalog(printedOldB006MaskToken) }

func printedOldB006MaskToken() TokenTemplate {
	return TokenTemplate{
		Slug: "mask",
		Card: game.Card{
			Name:     "Mask",
			TypeLine: "Token Enchantment — Aura",
			Colors:   []string{"W"},
		},
		Enchant:      TargetPermanent("permanent"),
		Replacements: []game.ReplacementEffect{UmbraArmor()},
		Text:         "Enchant permanent\nUmbra armor (If enchanted permanent would be destroyed, instead remove all damage from it and destroy this Aura.)",
	}
}

// OldB006UntapEachEnchantedPermanentYouControl is "untap each enchanted
// permanent you control": every tapped permanent of yours with an Aura
// attached (CR 303.4), read before any untaps.
func OldB006UntapEachEnchantedPermanentYouControl(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, YouControl()) {
		if c.Tapped && B18AurasAttachedTo(g, c.InstanceID) > 0 {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

func oldB006IsAuraCard(c game.Card) bool { return c.IsEnchantment() && c.HasSubtypeWord("Aura") }

// OldB006ReturnAurasTail is the second half of Estrid's −7, "then do the same
// for Aura cards": it runs after the non-Aura enchantments have all entered
// and any question they raised is answered, and returns every Aura card now in
// your graveyard. Each Aura chooses what it enchants as it enters (CR 303.4f);
// one that can't enchant anything stays in the graveyard.
var OldB006ReturnAurasTail = TogetherTail("old-b006/estrid-return-auras", func(ctx *Context, _ []uuid.UUID, _ int) error {
	return oldB006ReturnAuras(ctx)
})

func oldB006ReturnAuras(ctx *Context) error {
	ids := GraveyardCardIDs(ctx, ctx.Controller(), oldB006IsAuraCard)
	if len(ids) == 0 {
		return nil
	}
	return ReturnFromGraveyardTogether{Targets: ids, Controller: ctx.Controller()}.Apply(ctx)
}

// OldB006MillSevenReturnEnchantments is Estrid's −7: mill seven, return every
// non-Aura enchantment card from your graveyard to the battlefield, then every
// Aura card.
func OldB006MillSevenReturnEnchantments(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (MillCards{N: 7}).Apply(ctx); err != nil {
		return err
	}
	ids := GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
		return c.IsEnchantment() && !oldB006IsAuraCard(c)
	})
	if len(ids) == 0 {
		return oldB006ReturnAuras(ctx)
	}
	return ReturnFromGraveyardTogether{Targets: ids, Controller: ctx.Controller(), ThenTail: OldB006ReturnAurasTail}.Apply(ctx)
}
