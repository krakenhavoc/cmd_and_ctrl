package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// std_b43_helpers.go — shared predicates for Standard set batches 43-46.

// APermanentYouControlWasTurnedFaceUp is "whenever a face-down permanent
// you control is turned face up" (Sumala Sentry). Unlike
// ACreatureYouControlWasTurnedFaceUp it does not ask for a creature: a
// disguised land or artifact turns up too (CR 708.8: the permanent is
// face up by the time the event is emitted).
func APermanentYouControlWasTurnedFaceUp(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller
}

// Token rows for batches 43-46, added from this file's init so the shared
// table's column alignment is not rewritten (#4452). The keys read like the
// printed text, as the table's do.
func init() {
	for key, card := range map[string]game.Card{
		// Soul Search, Teysa, Opulent Oligarch.
		"1/1 white and black Spirit with flying": {Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1, Colors: []string{"W", "B"}, Keywords: []string{"flying"}},
		// Intrude on the Mind: the card's own Thopter is a 0/0 that takes counters.
		"0/0 colorless Thopter artifact with flying": {Name: "Thopter", TypeLine: "Token Artifact Creature — Thopter", Power: 0, Toughness: 0, Keywords: []string{"flying"}, PrintedPTKnown: true},
		// Tolsimir, Midnight's Light.
		"5/5 green and white Legendary Wolf named Voja Fenstalker": {Name: "Voja Fenstalker", TypeLine: "Token Legendary Creature — Wolf", Power: 5, Toughness: 5, Colors: []string{"G", "W"}, Keywords: []string{"trample"}},
	} {
		if _, dup := tokenTable[key]; dup {
			panic("std_b43_helpers: token key already in the table: " + key)
		}
		tokenTable[key] = card
	}
}
