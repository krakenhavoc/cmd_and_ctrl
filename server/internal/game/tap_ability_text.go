package game

import "strings"

// tap_ability_text.go — "has an activated ability with {T} in its cost"
// for a creature the catalog does not know (#3860, Magewright's Stone).
//
// The engine can see a catalogued creature's tap abilities
// (ActivatedAbilitiesForCard, ManaAbilitiesForCard), and nothing else:
// game.Card carries no oracle text, so an uncatalogued Llanowar-alike
// has no ability the engine could point at. The printed clause is
// about TEXT ("has an activated ability with {T} in its cost"), so the
// honest source for an uncatalogued creature is its text, read once.
//
// The deck importer stamps Card.PrintsTapAbility the same way it stamps
// PrintsManaAbility (#2258): the last place the Scryfall record is in
// scope. A multi-faced card is judged on the face that is up: SetFace
// re-derives the stamp from that face's own text (Face.OracleText), so
// a transformed werewolf is judged on the side that is up. Face-down
// permanents have no text (CR 708.2a), so none.

// PrintsTapAbilityText reports whether a permanent card's printed text
// has an activated ability whose cost includes {T}: a line, reminder
// text removed, shaped "<cost>: <effect>" whose cost names {T}.
//
// The line shape is the rules' own (CR 602.1): an activated ability is
// "[Cost]: [Effect]". What it deliberately refuses is a colon that is
// not that:
//
//   - inside quotation marks: "Creatures you control have '{T}: Add
//     {G}.'" and "Equipped creature has '{T}: ...'" GRANT the ability,
//     they are not the card's own;
//   - after a triggered ability's lead-in ("When", "Whenever", "At"):
//     what follows the colon is not an ability's effect;
//   - in a spell: an instant or sorcery has no abilities of a
//     permanent.
//
// texts are as for NeedsCatalogEffect: the card's text and each face's.
func PrintsTapAbilityText(typeLine string, texts ...string) bool {
	if typeLineHas(typeLine, "instant") || typeLineHas(typeLine, "sorcery") {
		return false
	}
	for _, text := range texts {
		if tapAbilityLineIn(text) {
			return true
		}
	}
	return false
}

// tapAbilityLineIn is PrintsTapAbilityText for one block of text.
func tapAbilityLineIn(text string) bool {
	if !strings.Contains(text, "{T}") {
		return false
	}
	for _, line := range strings.Split(stripReminderText(text), "\n") {
		cost, _, ok := strings.Cut(line, ":")
		if !ok || !strings.Contains(cost, "{T}") {
			continue
		}
		if strings.ContainsAny(cost, "\"“”") {
			continue
		}
		lead := strings.ToLower(strings.TrimSpace(cost))
		if strings.HasPrefix(lead, "when") || strings.HasPrefix(lead, "at ") || strings.HasPrefix(lead, "if ") {
			continue
		}
		return true
	}
	return false
}

// PrintedTapAbility reports whether the card's own printed text has an
// activated ability with {T} in its cost, as the face that is up prints
// it: the importer's stamp, which SetFace re-derives from the face's
// own text whenever a multi-faced card turns a face up. False for a
// face-down permanent, which has no text (CR 708.2a).
//
// It reads TEXT, so it is the answer for an uncatalogued card only;
// callers join it with the catalog's view of the card
// (ActivatedAbilitiesForCard / ManaAbilitiesForCard), which is the
// answer for a catalogued one and also covers what a granting
// permanent gave it.
func (c *Card) PrintedTapAbility() bool {
	return c != nil && c.PrintsTapAbility && !c.FaceDownIsPermanent()
}
