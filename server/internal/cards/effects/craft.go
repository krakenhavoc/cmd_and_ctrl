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
// The material clauses, one constructor per printed shape:
//
//	"Craft with artifact"                      CraftWith
//	"Craft with two creatures"                 CraftWithN
//	"Craft with Island"                        CraftWithSubtype
//	"Craft with one or more creatures"         CraftWithOneOrMore
//	"Craft with one or more Dinosaurs"         CraftWithOneOrMoreSubtype
//	"Craft with one or more"                   CraftWithOneOrMore("")
//	"Craft with two that share a card type"    CraftWithTwoSharingACardType
//	"Craft with a Dinosaur, a Merfolk, …"      CraftWithEachOf
//	"Craft with four or more red instant
//	 and/or sorcery cards"                     CraftWithCardsOrMore
//
// The last four are ADR 0137's 2026-10-10 amendment (#2709): an open
// count (ExilePermanentsCost.OrMore), a rule over the chosen set
// (ShareCardType, EachSubtype) and a graveyard-only clause
// (GraveyardOnly).

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

// CraftWithOneOrMore is "Craft with one or more <card type>s" —
// Altar of the Wretched's and Paleontologist's Pick-Axe's "one or more
// creatures" — and, with an empty card type, Sunbird Standard's bare
// "Craft with one or more": any other permanents you control and/or
// any cards in your graveyard. The activator names how many (CR
// 602.2b), at least one.
func CraftWithOneOrMore(cardType string) game.AbilityCost {
	label := "one or more other permanents you control and/or cards from your graveyard"
	if cardType != "" {
		label = "one or more " + cardType + "s you control and/or " + cardType + " cards from your graveyard"
	}
	cost := craftMaterials(1, cardType, "", label)
	cost.ExilePermanents.OrMore = true
	return cost
}

// CraftWithOneOrMoreSubtype is "Craft with one or more <subtype>s" —
// Saheeli's Lattice's "one or more Dinosaurs".
func CraftWithOneOrMoreSubtype(subtype string) game.AbilityCost {
	cost := craftMaterials(1, "", subtype,
		"one or more "+subtype+"s you control and/or "+subtype+" cards from your graveyard")
	cost.ExilePermanents.OrMore = true
	return cost
}

// CraftWithTwoSharingACardType is Eye of Ojer Taq's "Craft with two
// that share a card type": any two from among other permanents you
// control and/or cards in your graveyard, as long as some card type is
// on both.
func CraftWithTwoSharingACardType() game.AbilityCost {
	cost := craftMaterials(2, "", "",
		"two that share a card type, from among other permanents you control and/or cards from your graveyard")
	cost.ExilePermanents.ShareCardType = true
	return cost
}

// CraftWithEachOf is "Craft with a Dinosaur, a Merfolk, a Pirate, and
// a Vampire" (Throne of the Grim Captain): one material for each
// subtype, from among permanents you control and/or cards in your
// graveyard, none of them filling two.
func CraftWithEachOf(subtypes ...string) game.AbilityCost {
	parts := make([]string, len(subtypes))
	for i, st := range subtypes {
		article := "a"
		if strings.ContainsRune("AEIOU", rune(st[0])) {
			article = "an"
		}
		parts[i] = article + " " + st
	}
	label := strings.Join(parts, ", ")
	if len(parts) > 1 {
		label = strings.Join(parts[:len(parts)-1], ", ") + ", and " + parts[len(parts)-1]
	}
	cost := craftMaterials(len(subtypes), "", "", label+", from among permanents you control and/or cards in your graveyard")
	cost.ExilePermanents.EachSubtype = append([]string(nil), subtypes...)
	return cost
}

// CraftWithCardsOrMore is a graveyard-only material clause with a
// floor: Ore-Rich Stalactite's "four or more red instant and/or sorcery
// cards" is CraftWithCardsOrMore(4, "R", "red", "instant", "sorcery").
// `color` is the one-letter code ("" for none) and `colorWord` how the
// label says it; the card types are "and/or".
func CraftWithCardsOrMore(n int, color, colorWord string, cardTypes ...string) game.AbilityCost {
	quality := strings.Join(cardTypes, " and/or ")
	if colorWord != "" {
		quality = colorWord + " " + quality
	}
	cost := craftMaterials(n, "", "", craftCountWord(n)+" or more "+quality+" cards from your graveyard")
	ec := cost.ExilePermanents
	ec.OrMore = true
	ec.GraveyardOnly = true
	ec.Color = color
	ec.CardTypes = append([]string(nil), cardTypes...)
	return cost
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

// craftMaterialsOf is CR 702.167c's link read off the permanent itself,
// for a static ability whose source is the crafted permanent (a
// characteristic-defining ability, a keyword grant, a mana ability):
// the cards still in exile as its materials. Nil for a permanent no
// craft ability made.
//
// Caller holds g.mu (a layer pass or a mana ability does).
func craftMaterialsOf(g *game.Game, source *game.Card) []game.Card {
	if source == nil || len(source.CraftedWith) == 0 {
		return nil
	}
	return g.CraftMaterialsForEffect(source.CraftedWith)
}

// craftMaterialsTotalPower is "the total power of the exiled cards used
// to craft it" (Mastercraft Raptor, Wretched Bonemass): each card's
// power where it is, its own characteristic-defining ability applied
// (the rulings' Souls of the Lost), a negative total read as it is.
func craftMaterialsTotalPower(g *game.Game, source *game.Card) int {
	total := 0
	for _, c := range craftMaterialsOf(g, source) {
		total += g.PowerAnywhereForEffect(c)
	}
	return total
}

// craftMaterialColors is "each color among the exiled cards used to
// craft it" (Sunbird Effigy): the colours the materials have, in WUBRG
// order, each once.
func craftMaterialColors(materials []game.Card) []string {
	var out []string
	for _, col := range []string{"W", "U", "B", "R", "G"} {
		for _, c := range materials {
			if c.HasColor(col) {
				out = append(out, col)
				break
			}
		}
	}
	return out
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
