package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// craft.go — CR 702.167, craft (ADR 0137).
//
//	"Craft with [materials] [cost]" means "[Cost], Exile this permanent,
//	 Exile [materials] from among permanents you control and/or cards in
//	 your graveyard: Return this card to the battlefield transformed
//	 under its owner's control. Activate only as a sorcery."
//
// The whole keyword is one constructor, Craft, because each half is a
// rule a hand-written ability would get wrong:
//
//   - The materials are the exile-a-permanent component with its
//     graveyard half switched on (game.ExilePermanentsCost.FromGraveyard,
//     CR 702.167b). A material named without the word "card" may be a
//     permanent you control or a card in your graveyard, mixed freely.
//     The source is never one of them: it pays "Exile this permanent"
//     (CR 118.3), so the clause always excludes it.
//   - The return is game.ReturnCraftedFromExileForEffect: the card
//     the cost exiled comes back on its back face as a NEW object, under
//     its OWNER's control whoever activated it, and remembers what was
//     exiled to make it (CR 702.167c, Card.CraftedWith).
//   - "Activate only as a sorcery" is SorcerySpeed.
//
// Material clauses with a fixed count and one quality are what this
// covers: "Craft with artifact" (CraftWith), "Craft with two creatures"
// (CraftWithN), "Craft with Island" (CraftWithSubtype). Still out of
// scope (ADR 0137): "one or more" (an announced count), "two that share
// a card type" and "a Dinosaur, a Merfolk, a Pirate, and a Vampire" (a
// set rule over the picks), and "four or more red instant and/or
// sorcery cards" (graveyard-only, a variable count).

// Craft is the craft keyword ability (CR 702.167a). `label` is the
// printed keyword line without its reminder text — "Craft with
// artifact {3}{W}{W}" — which the oracle check matches; `mana` is the
// mana half of the cost; `materials` is one of the CraftWith
// constructors below.
func Craft(label, mana string, materials game.AbilityCost) ActivatedAbility {
	return ActivatedAbility{
		Label:        label,
		Cost:         Plus(ManaCost(mana), ExileThis(), materials),
		SorcerySpeed: true,
		Effect:       returnCraftedCard,
	}
}

// CraftWith is "Craft with <card type>" — one material: "another
// artifact you control or an artifact card from your graveyard",
// "a creature you control or a creature card from your graveyard".
// `cardType` is a permanent card type, lowercase ("artifact",
// "creature", "land").
func CraftWith(cardType string) game.AbilityCost {
	return craftMaterials(1, cardType, "", craftLabel(1, cardType))
}

// CraftWithN is "Craft with <n> <card type>s" — Visage of Dread's
// "two creatures", Tetzin's "six artifacts".
func CraftWithN(n int, cardType string) game.AbilityCost {
	return craftMaterials(n, cardType, "", craftLabel(n, cardType))
}

// CraftWithSubtype is "Craft with <subtype>" — Waterlogged Hulk's
// "Island", Kaslem's Stonetree's "Cave": one permanent you control or
// one card in your graveyard with that subtype, of any card type.
func CraftWithSubtype(subtype string) game.AbilityCost {
	article := "a"
	if strings.ContainsRune("AEIOU", rune(subtype[0])) {
		article = "an"
	}
	return craftMaterials(1, "", subtype,
		article+" "+subtype+" you control or "+article+" "+subtype+" card from your graveyard")
}

func craftMaterials(n int, cardType, subtype, label string) game.AbilityCost {
	return game.AbilityCost{ExilePermanents: &game.ExilePermanentsCost{
		Count:         n,
		CardType:      cardType,
		Subtype:       subtype,
		ExcludeSource: true,
		FromGraveyard: true,
		Label:         label,
	}}
}

// craftLabel is the material clause as the reminder text words it,
// shown above the client's picker.
func craftLabel(n int, cardType string) string {
	if n == 1 {
		article := "a"
		if strings.ContainsRune("aeiou", rune(cardType[0])) {
			article = "an"
		}
		// Every craft card is an artifact, so its reminder text says
		// "another artifact" and plain "a creature".
		self := article
		if cardType == "artifact" {
			self = "another"
		}
		return self + " " + cardType + " you control or " + article + " " + cardType + " card from your graveyard"
	}
	return craftCountWord(n) + " from among " + cardType + "s you control and/or " + cardType + " cards in your graveyard"
}

func craftCountWord(n int) string {
	words := []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}
	if n >= 0 && n < len(words) {
		return words[n]
	}
	return "several"
}

// returnCraftedCard is every craft ability's effect: the card the cost
// exiled — still under the source's instance ID, which an exile keeps —
// returns transformed, linked to the materials the same cost exiled.
// A package-level func, so the ability captures nothing (undo safety).
func returnCraftedCard(g *game.Game, item *game.StackItem) error {
	_, err := g.ReturnCraftedFromExileForEffect(item.SourceCardID, NewContext(g, item).Paid().Exiled)
	return err
}

// CraftMaterials is CR 702.167c's "the exiled cards used to craft it",
// read by an ability of the crafted permanent: the cards still in exile
// as the materials its craft ability exiled, as value copies. Read
// through the source OBJECT, so a trigger that resolves after the
// permanent has left reads its last-known link (CR 608.2h). Nil when
// the source was not crafted.
func CraftMaterials(ctx *Context) []game.Card {
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	return ctx.Game.CraftMaterialsForEffect(info.CraftedWith)
}
