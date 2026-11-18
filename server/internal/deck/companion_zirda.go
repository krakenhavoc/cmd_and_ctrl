package deck

import (
	"regexp"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// companion_zirda.go — Zirda, the Dawnwaker's companion condition
// (ADR 0172 §9, #5909): "Each permanent card in your starting deck has
// an activated ability."
//
// The check reads each card's printed characteristics outside the game
// (zirdaFaces) and asks two questions of them:
//
//   - is it a permanent card? Artifact, battle, creature, enchantment,
//     land or planeswalker (CR 110.4). An instant or sorcery, kindred
//     ones included, needs nothing;
//   - if so, does it have an activated ability (printsActivatedAbility)?
//
// An activated ability is "[Cost]: [Effect]" (CR 602.1, 113.3b), which
// takes in mana abilities (CR 605.1a), loyalty abilities (CR 606) and
// the ability-word forms (boast, exhaust, channel, forecast...), all
// printed with their colon. The keyword abilities that are activated
// (equip, cycling, ninjutsu...) print without one, the colon being in
// their reminder text, so they are listed (activatedKeywords). And a
// land with a basic land type has "{T}: Add [mana]" even when its text
// box doesn't say so (CR 305.6).
//
// What it refuses: a colon in reminder text (a Treasure's "It's an
// artifact with ..."), one inside quotation marks (an ability the card
// GRANTS, "Equipped creature has '{T}: ...'", is not its own) and one
// after a triggered ability's lead-in.

// zirdaOracle is Zirda, the Dawnwaker's oracle ID.
const zirdaOracle = "7b515c56-7b1e-42a8-82b0-1b30201fb162"

func init() {
	companionConditions[zirdaOracle] = companionCondition{
		Text:      "each permanent card in your starting deck must have an activated ability",
		Offenders: offendersEach(permanentWithoutActivatedAbility),
	}
}

// zirdaFace is the part of a card's printed characteristics the
// condition reads.
type zirdaFace struct {
	Name, TypeLine, OracleText string
}

// zirdaFaces is what a card in a deck, which is outside the game,
// prints, as companionManaCosts reads its cost:
//
//   - a split card is both halves combined (CR 709.4);
//   - a double-faced card has only its front face (CR 712.8a);
//   - an adventurer, flip or preparation card has only its normal
//     characteristics (CR 715.4, 710.2, 722.4), which Scryfall writes
//     as the first face.
func zirdaFaces(c cards.Card) []zirdaFace {
	if len(c.CardFaces) == 0 {
		return []zirdaFace{{c.Name, c.TypeLine, c.OracleText}}
	}
	faces := c.CardFaces
	if c.Layout != "split" {
		faces = faces[:1]
	}
	out := make([]zirdaFace, 0, len(faces))
	for _, f := range faces {
		out = append(out, zirdaFace{f.Name, f.TypeLine, f.OracleText})
	}
	return out
}

// permanentWithoutActivatedAbility is Zirda's per-card test: true for a
// permanent card none of whose faces has an activated ability.
func permanentWithoutActivatedAbility(c cards.Card) bool {
	faces := zirdaFaces(c)
	permanent := false
	for _, f := range faces {
		if zirdaIsPermanentTypeLine(f.TypeLine) {
			permanent = true
		}
	}
	if !permanent {
		return false
	}
	for _, f := range faces {
		if zirdaHasBasicLandType(f.TypeLine) || printsActivatedAbility(f.Name, f.OracleText) {
			return false
		}
	}
	return true
}

// zirdaPermanentTypes are the six permanent card types (CR 110.4).
var zirdaPermanentTypes = map[string]bool{
	"artifact": true, "battle": true, "creature": true,
	"enchantment": true, "land": true, "planeswalker": true,
}

// zirdaIsPermanentTypeLine reports whether a face's card types, the words
// before the em dash, include a permanent type.
func zirdaIsPermanentTypeLine(typeLine string) bool {
	types, _, _ := strings.Cut(typeLine, "—")
	for _, w := range strings.Fields(strings.ToLower(types)) {
		if zirdaPermanentTypes[w] {
			return true
		}
	}
	return false
}

// zirdaBasicLandTypes are the subtypes that give a land "{T}: Add [mana]"
// (CR 305.6).
var zirdaBasicLandTypes = map[string]bool{
	"plains": true, "island": true, "swamp": true, "mountain": true, "forest": true,
}

// zirdaHasBasicLandType reports whether a face is a land with a basic land
// type, which has an intrinsic mana ability (CR 305.6).
func zirdaHasBasicLandType(typeLine string) bool {
	types, subtypes, ok := strings.Cut(typeLine, "—")
	if !ok || !strings.Contains(strings.ToLower(types), "land") {
		return false
	}
	for _, w := range strings.Fields(strings.ToLower(subtypes)) {
		if zirdaBasicLandTypes[w] {
			return true
		}
	}
	return false
}

// activatedKeywords are the keyword abilities that are activated
// abilities and print without a colon, from CR 702: equip (702.6a),
// cycling (702.29a; typecycling, 702.29e, is matched by
// typecyclingRE), ninjutsu (702.49a) and commander ninjutsu (702.49d),
// transmute (702.53a), aura swap (702.65a), fortify (702.67a),
// transfigure (702.71a), reinforce (702.77a), unearth (702.84a), level
// up (702.87a), scavenge (702.97a), outlast (702.107a), crew
// (702.122a), embalm (702.128a), eternalize (702.129a), encore
// (702.141a), reconfigure (702.151a), craft (702.167a), saddle
// (702.171a) and station (702.184a). The ones written "Keyword —
// [Cost]: [Effect]" (forecast, boast, exhaust, power-up, solved) carry
// their colon and need no entry. Morph, foretell, plot, suspend and
// companion are special actions or static abilities, not activated
// ones.
var activatedKeywords = []string{
	"equip", "cycling", "ninjutsu", "commander ninjutsu", "transmute",
	"aura swap", "fortify", "transfigure", "reinforce", "unearth",
	"level up", "scavenge", "outlast", "crew", "embalm", "eternalize",
	"encore", "reconfigure", "craft", "saddle", "station",
}

// typecyclingRE is a typecycling keyword: "Plainscycling {2}", "Basic
// landcycling {1}{G}", "Wizardcycling {3}" (CR 702.29e).
var typecyclingRE = regexp.MustCompile(`^(basic )?[a-z]+cycling(\s|$)`)

// printsActivatedAbility reports whether one face's oracle text has an
// activated ability of its own.
func printsActivatedAbility(name, text string) bool {
	if name != "" {
		// A name with a colon in it ("Vault 101: Birthday Party") is
		// not a cost.
		text = strings.ReplaceAll(text, name, "~")
	}
	for _, line := range strings.Split(zirdaStripReminder(text), "\n") {
		if activatedAbilityLine(line) || activatedKeywordLine(line) {
			return true
		}
	}
	return false
}

// activatedAbilityLine reports whether a line, reminder text removed,
// reads "[Cost]: [Effect]" (CR 602.1) as the card's own ability.
func activatedAbilityLine(line string) bool {
	cost, _, ok := strings.Cut(line, ":")
	if !ok || strings.ContainsAny(cost, "\"“”") {
		return false
	}
	// An ability word or keyword before an em dash ("Boast — {1}{R}",
	// "Landfall — Whenever ...") is not the cost's first word.
	if i := strings.LastIndex(cost, "—"); i >= 0 {
		cost = cost[i+len("—"):]
	}
	lead := strings.ToLower(strings.TrimSpace(cost))
	if lead == "" {
		return false
	}
	for _, trigger := range []string{"when ", "whenever ", "at ", "if "} {
		if strings.HasPrefix(lead, trigger) {
			return false
		}
	}
	return true
}

// activatedKeywordLine reports whether a keyword line names one of the
// activated keywords: "Equip {2}", "Equip legendary creature {3}",
// "Crew 3", "Station", "Flying, cycling {2}". A sentence about the
// keyword ("Equip abilities you activate cost {1} less") is not one.
func activatedKeywordLine(line string) bool {
	for _, part := range strings.Split(strings.ReplaceAll(line, ";", ","), ",") {
		part = strings.ToLower(strings.TrimSpace(part))
		if typecyclingRE.MatchString(part) && !strings.Contains(part, "abilit") {
			return true
		}
		for _, kw := range activatedKeywords {
			rest, ok := strings.CutPrefix(part, kw)
			if !ok {
				continue
			}
			if rest != "" && !strings.HasPrefix(rest, " ") && !strings.HasPrefix(rest, "—") {
				continue // "equipped", "equipment"
			}
			if strings.Contains(rest, "abilit") {
				continue
			}
			return true
		}
	}
	return false
}

// zirdaStripReminder removes parenthesised reminder text.
func zirdaStripReminder(text string) string {
	if !strings.ContainsRune(text, '(') {
		return text
	}
	var b strings.Builder
	depth := 0
	for _, r := range text {
		switch {
		case r == '(':
			depth++
		case r == ')' && depth > 0:
			depth--
		case depth == 0:
			b.WriteRune(r)
		}
	}
	return b.String()
}
