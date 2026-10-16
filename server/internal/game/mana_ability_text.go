package game

import (
	"regexp"
	"strings"
)

// mana_ability_text.go — "a card with a mana ability" read off the
// card's printed text (#3943, Moonsilver Key).
//
// "Search your library for an artifact card with a mana ability" asks
// what the card in the library HAS, which is what it prints (CR 109.3:
// an object's abilities are part of its characteristics; a card in a
// library has nothing else granting it one). The engine's other view of
// a card's abilities is the catalog, which knows only the cards it runs,
// and Scryfall's `produced_mana` is not the answer either: it is set for
// any card that adds mana, Su-Chi's dies trigger, a Treasure in reminder
// text and an Equipment's grant included. So the importer reads the
// printed text once, by the rules' own definition, and stamps the
// answer on the card (Card.ManaAbilityText), the way it stamps
// PrintsTapAbility (#3860). game.Card carries no oracle text.
//
// CR 605.1a: an ACTIVATED ability is a mana ability if it doesn't
// require a target, could add mana when it resolves, isn't a loyalty
// ability, and its cost and effect don't move any card to or from a
// library. CR 605.1b: a TRIGGERED ability is one if it doesn't require a
// target, triggers from activating or resolving an activated mana
// ability or from mana being added, and could add mana. ManaAbilityInText
// recognises both shapes; PrintsManaText (#2258), by contrast, flags any
// "Add" clause, which is what the "may make mana the engine doesn't
// run" warning wants and not what the rules call a mana ability.

// PrintedFact is what a card's printed text says about one question,
// as the deck importer read it. The zero value is "not read": a card
// that never went through import (a token, a test fixture, a restore
// point written before the stamp existed), whose answer has to come
// from somewhere else.
type PrintedFact uint8

const (
	// PrintedUnknown: the text was never read.
	PrintedUnknown PrintedFact = iota
	// PrintedNo: the text was read and says no.
	PrintedNo
	// PrintedYes: the text was read and says yes.
	PrintedYes
)

// PrintedFactOf is the fact for a text that was read.
func PrintedFactOf(yes bool) PrintedFact {
	if yes {
		return PrintedYes
	}
	return PrintedNo
}

var (
	// manaAbilityAddRE finds "could add mana" in lower-cased effect
	// text: "add {g}", "add one mana of any color", "add x mana", "adds
	// an additional {r}", "add an amount of {g}", "add that much mana",
	// and Doubling Cube's "double the amount of each type of unspent
	// mana you have", which adds mana too.
	manaAbilityAddRE = regexp.MustCompile(`\badds? (\{|one |two |three |four |five |six |seven |eight |nine |ten |x |an amount|that much|mana|[a-z]+ mana|an additional|that )|\bdoubles? the amount of [^.]*\bmana\b`)
	// manaAbilityLibraryRE is a cost or effect that moves a card to or
	// from a library (CR 605.1a's last criterion): anything naming a
	// library, and the keyword actions that take a card out of one.
	// Scry keeps the cards in the library, so it is not here.
	manaAbilityLibraryRE = regexp.MustCompile(`\blibrar(y|ies)\b|\bmill(s|ed)?\b|\bdraws?\b|\bsurveil\b|\bexplores?\b|\bdiscover\b|\bmanifest\b`)
	// manaAbilityTargetRE is CR 115.6's "target": an ability that
	// requires one is not a mana ability (CR 605.5a).
	manaAbilityTargetRE = regexp.MustCompile(`\btargets?\b`)
	// manaEventRE is a trigger condition that is a mana event (CR
	// 605.1b): something tapped for mana ("whenever a player taps a
	// land for mana", "whenever you tap a permanent for {c}"), a mana
	// ability, or mana being added ("whenever a land's ability causes
	// you to add one or more mana").
	manaEventRE = regexp.MustCompile(`\btap(s|ped)?\b[^,]*\bfor\b|\bmana abilit|\badds?\b[^,]*\bmana\b`)
	// loyaltyCostRE is a loyalty ability's cost (CR 606): "+1", "−2",
	// "0", "−X", Scryfall's minus sign or a hyphen.
	loyaltyCostRE = regexp.MustCompile(`^\[?[+\x{2212}-]?(\d+|x)\]?$`)
	// tableRowRE is a line that continues the ability above it: a die
	// roll's result row ("2 — Add {W}.", "1-9 | …") or a mode bullet.
	tableRowRE = regexp.MustCompile(`^\s*(\d+(\s*[\x{2014}\x{2013}-]\s*\d+)?\s*[\x{2014}|]|\x{2022})`)
	// sentenceRE splits an effect into sentences.
	sentenceRE = regexp.MustCompile(`\.\s+`)
)

// ManaAbilityInText reports whether a card's printed text has a mana
// ability (CR 605.1a or 605.1b). typeLine and texts are the card's as
// the deck importer has them: the card's text and each face's, reminder
// text included (it is removed here). An instant or sorcery has none
// (CR 605.5b), and a land with a basic land type has its intrinsic one
// (CR 305.6) whatever the box says.
//
// What it reads, line by line, after reminder text is removed:
//
//   - an activated ability, "[cost]: [effect]" (CR 602.1), whose cost
//     is not a loyalty cost and whose effect adds mana, names no target
//     and moves no card to or from a library. A colon inside quotation
//     marks is a granted ability, not the card's own ("Equipped creature
//     has '{T}: Add …'"), and so is quoted text inside an effect (a
//     Treasure the effect makes). A sentence of the effect that creates
//     a delayed triggered ability ("When that mana is spent …", "Draw a
//     card at the beginning of the next turn's upkeep") is that ability,
//     not this one's effect (Pyromancer's Goggles, Barbed Sextant).
//   - a triggered ability, "Whenever [a mana event], [effect]", whose
//     effect adds mana and names no target (Gauntlet of Power, Caged
//     Sun). "Whenever you cast a spell, add {R}" and "When this dies,
//     add {C}" trigger on something else and are not mana abilities
//     (CR 605.5a, Su-Chi).
//
// It errs toward "no" on a shape it does not read, which for Moonsilver
// Key is the weaker search.
func ManaAbilityInText(typeLine string, texts ...string) bool {
	if typeLineHas(typeLine, "instant") || typeLineHas(typeLine, "sorcery") {
		return false
	}
	if printsBasicLandType(typeLine) {
		return true
	}
	for _, text := range texts {
		for _, line := range abilityLines(stripReminderText(text)) {
			if manaAbilityLine(line) {
				return true
			}
		}
	}
	return false
}

// printsBasicLandType reports a land whose type line names a basic land
// type, which gives it that type's intrinsic mana ability (CR 305.6).
func printsBasicLandType(typeLine string) bool {
	head, subtypes, ok := strings.Cut(typeLine, "—")
	if !ok || !typeLineHas(head, "land") {
		return false
	}
	for _, st := range strings.Fields(strings.ToLower(subtypes)) {
		switch st {
		case "plains", "island", "swamp", "mountain", "forest":
			return true
		}
	}
	return false
}

// abilityLines splits text into one entry per ability: a die roll's
// result rows and a mode's bullets join the line above them, so Jack-in-
// the-Mox's "{T}: Roll a six-sided die" is read with the "2 — Add {W}"
// it can lead to.
func abilityLines(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		if len(out) > 0 && tableRowRE.MatchString(line) {
			out[len(out)-1] += " " + strings.TrimSpace(line)
			continue
		}
		out = append(out, line)
	}
	return out
}

// manaAbilityLine is ManaAbilityInText for one ability's line.
func manaAbilityLine(line string) bool {
	low := strings.ToLower(strings.TrimSpace(line))
	if low == "" {
		return false
	}
	if rest, ok := strings.CutPrefix(low, "whenever "); ok {
		cond, effect, ok := strings.Cut(rest, ",")
		if !ok || !manaEventRE.MatchString(cond) {
			return false
		}
		e := ownEffect(effect)
		return manaAbilityAddRE.MatchString(e) && !manaAbilityTargetRE.MatchString(e)
	}
	if strings.HasPrefix(low, "when ") || strings.HasPrefix(low, "at ") {
		return false
	}
	cost, effect, ok := strings.Cut(low, ":")
	if !ok || strings.ContainsAny(cost, "\"“”") {
		return false
	}
	cost = strings.TrimSpace(cost)
	if loyaltyCostRE.MatchString(strings.ReplaceAll(cost, " ", "")) {
		return false
	}
	if strings.HasPrefix(cost, "if ") || strings.HasPrefix(cost, "as long as") || strings.HasPrefix(cost, "for as long") {
		return false
	}
	e := ownEffect(effect)
	if !manaAbilityAddRE.MatchString(e) {
		return false
	}
	if manaAbilityTargetRE.MatchString(e) || manaAbilityTargetRE.MatchString(cost) {
		return false
	}
	return !manaAbilityLibraryRE.MatchString(e) && !manaAbilityLibraryRE.MatchString(cost)
}

// ownEffect is an ability's own effect text: quoted text removed (a
// granted ability, a token's ability), and every sentence that creates a
// delayed triggered ability dropped — from the first sentence that
// starts "When" or "Whenever" on, and any sentence that waits for "the
// beginning of" a step. Lower-cased input.
func ownEffect(effect string) string {
	effect = stripQuoted(effect)
	var kept []string
	for _, s := range sentenceRE.Split(effect, -1) {
		s = strings.TrimSpace(s)
		if strings.HasPrefix(s, "when ") || strings.HasPrefix(s, "whenever ") {
			break
		}
		if strings.Contains(s, "at the beginning of") {
			continue
		}
		kept = append(kept, s)
	}
	return strings.Join(kept, ". ")
}

// stripQuoted removes every span in straight or curly double quotes.
func stripQuoted(s string) string {
	var b strings.Builder
	in := false
	for _, r := range s {
		switch r {
		case '“':
			in = true
		case '”':
			in = false
		case '"':
			in = !in
		default:
			if !in {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// HasPrintedManaAbility answers "a card with a mana ability" (CR 605.1)
// for Moonsilver Key's search: what the card prints.
//
// The importer's stamp is the answer whenever the text was read,
// catalogued card or not. The catalog is only the fallback for a card
// whose text was never read (a test fixture, a restore point older than
// the stamp): it is a statement about what the engine RUNS, which for a
// handful of cards is not what the rules call a mana ability (Millikin's
// "{T}, Mill a card: Add {C}" moves a card from a library, so CR 605.1a
// makes it an ordinary activated ability). A face-down card has no text
// (CR 708.2a) and so no mana ability.
func HasPrintedManaAbility(c Card) bool {
	if c.FaceDownIsPermanent() {
		return false
	}
	switch c.ManaAbilityText {
	case PrintedYes:
		return true
	case PrintedNo:
		return false
	}
	return len(ManaAbilitiesForCard(c)) > 0 || len(ManaTriggersForCard(c)) > 0
}
