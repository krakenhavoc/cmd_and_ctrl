package deck

import (
	"slices"
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// companion_s82_rpr_b02.go — the companion conditions of Lurrus, Kaheera,
// Keruga, Lutri and Obosh (ADR 0172 §9, S82 reprint batch 2, #5908). Each is
// registered from here so that a card batch adds its condition without
// touching the registry literal in companion.go.
//
// They read printed characteristics only, outside the game: a
// double-faced or modal card is its front face (CR 712.8a), a split card's
// mana value is both halves combined (CR 708.4 and CR 202.3), and X is zero
// outside the stack (CR 202.3e). A land card has mana value 0 and fulfils
// every "or land cards" clause by being a land.

func init() {
	// Lurrus of the Dream-Den: "Each permanent card in your starting deck
	// has mana value 2 or less."
	companionConditions["3bc757c1-3adb-4321-8832-8e1cc9e687f7"] = companionCondition{
		Text: "each permanent card in your starting deck must have mana value 2 or less",
		Offenders: offendersEach(func(c cards.Card) bool {
			return companionIsPermanentCard(c) && companionManaValue(c) > 2
		}),
	}
	// Kaheera, the Orphanguard: "Each creature card in your starting deck is
	// a Cat, Elemental, Nightmare, Dinosaur, or Beast card." A changeling is
	// every creature type (CR 702.73a), so it passes.
	companionConditions["35428c77-9731-4d4d-8f75-e939a491cf86"] = companionCondition{
		Text: "each creature card in your starting deck must be a Cat, Elemental, Nightmare, Dinosaur, or Beast",
		Offenders: offendersEach(func(c cards.Card) bool {
			if !strings.Contains(companionFrontTypeTypes(c), "Creature") || companionIsChangeling(c) {
				return false
			}
			for _, sub := range companionSubtypes(c) {
				if slices.Contains(kaheeraTypes, sub) {
					return false
				}
			}
			return true
		}),
	}
	// Keruga, the Macrosage: "Your starting deck contains only cards with
	// mana value 3 or greater and land cards."
	companionConditions["ca5121b2-e6d9-40bc-a2d2-05852c4efbe8"] = companionCondition{
		Text: "your starting deck may contain only cards with mana value 3 or greater and land cards",
		Offenders: offendersEach(func(c cards.Card) bool {
			return !companionIsLand(c) && companionManaValue(c) < 3
		}),
	}
	// Lutri, the Spellchaser: "Each nonland card in your starting deck has a
	// different name."
	companionConditions["158a6225-a246-4fd6-aa57-0df8067b4383"] = companionCondition{
		Text:      "each nonland card in your starting deck must have a different name",
		Offenders: lutriOffenders,
	}
	// Obosh, the Preypiercer: "Your starting deck contains only cards with
	// odd mana values and land cards."
	companionConditions["6985c8a9-9299-461b-8bef-bd2b2c739c11"] = companionCondition{
		Text: "your starting deck may contain only cards with odd mana values and land cards",
		Offenders: offendersEach(func(c cards.Card) bool {
			return !companionIsLand(c) && companionManaValue(c)%2 == 0
		}),
	}
}

// kaheeraTypes are the creature types Kaheera's condition allows.
var kaheeraTypes = []string{"Cat", "Elemental", "Nightmare", "Dinosaur", "Beast"}

// companionFrontTypeLine is the card's type line outside the game: the front
// face's (CR 712.8a), which Scryfall joins to the back's with " // ".
func companionFrontTypeLine(c cards.Card) string {
	line := c.TypeLine
	if line == "" && len(c.CardFaces) > 0 {
		line = c.CardFaces[0].TypeLine
	}
	front, _, _ := strings.Cut(line, " // ")
	return front
}

// companionFrontTypeTypes is the part of the type line before the dash: the
// supertypes and card types.
func companionFrontTypeTypes(c cards.Card) string {
	types, _, _ := strings.Cut(companionFrontTypeLine(c), " — ")
	return types
}

// companionSubtypes is the subtypes after the dash.
func companionSubtypes(c cards.Card) []string {
	_, subs, ok := strings.Cut(companionFrontTypeLine(c), " — ")
	if !ok {
		return nil
	}
	return strings.Fields(subs)
}

// companionIsLand reports a land card (CR 205.2a), a land by its front face.
func companionIsLand(c cards.Card) bool {
	return strings.Contains(companionFrontTypeTypes(c), "Land")
}

// companionIsPermanentCard reports a permanent card: an artifact, creature,
// enchantment, land, planeswalker or battle card (CR 110.4).
func companionIsPermanentCard(c cards.Card) bool {
	types := companionFrontTypeTypes(c)
	for _, t := range []string{"Artifact", "Creature", "Enchantment", "Land", "Planeswalker", "Battle"} {
		if strings.Contains(types, t) {
			return true
		}
	}
	return false
}

// companionIsChangeling reports a card with changeling on its front face.
func companionIsChangeling(c cards.Card) bool {
	if slices.Contains(c.Keywords, "Changeling") {
		return true
	}
	text := c.OracleText
	if text == "" && len(c.CardFaces) > 0 {
		text = c.CardFaces[0].OracleText
	}
	for _, line := range strings.Split(text, "\n") {
		if strings.EqualFold(strings.TrimSpace(line), "changeling") {
			return true
		}
	}
	return false
}

// companionManaValue is the card's mana value outside the game (CR 202.3):
// the symbols of its printed cost (companionManaCosts), X as zero, a hybrid
// symbol as its largest part ({2/W} is 2, {G/W} is 1) and a Phyrexian symbol
// as 1. A split card's halves are added.
func companionManaValue(c cards.Card) int {
	total := 0
	for _, cost := range companionManaCosts(c) {
		for _, sym := range manaSymbolRE.FindAllString(cost, -1) {
			total += companionSymbolValue(strings.Trim(sym, "{}"))
		}
	}
	return total
}

func companionSymbolValue(sym string) int {
	if n, err := strconv.Atoi(sym); err == nil {
		return n
	}
	switch sym {
	case "X", "Y", "Z":
		return 0
	}
	if first, _, ok := strings.Cut(sym, "/"); ok {
		if n, err := strconv.Atoi(first); err == nil {
			return n
		}
	}
	return 1
}

// lutriOffenders names each nonland card whose name appears on another
// nonland card of the starting deck, once each, in deck order.
func lutriOffenders(starting []cards.Card) []string {
	count := map[string]int{}
	for _, c := range starting {
		if !companionIsLand(c) {
			count[c.Name]++
		}
	}
	var out []string
	for _, c := range starting {
		if companionIsLand(c) || count[c.Name] < 2 || slices.Contains(out, c.Name) {
			continue
		}
		out = append(out, c.Name)
	}
	return out
}
