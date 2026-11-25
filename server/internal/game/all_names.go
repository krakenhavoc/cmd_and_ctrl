package game

import "strings"

// all_names.go — #6257: an object with "all names of nonlegendary
// creature cards in addition to its name" (Spy Kit's equipped creature).
//
// CR 201.2a: two objects have the same name if they have at least one
// name in common, even if one of them has additional names. So the
// equipped creature matches "a card named Grizzly Bears", and a Pithing
// Needle naming any nonlegendary creature card reaches it.
//
// The set is every nonlegendary creature card in the card reference,
// which the engine does not hold: the server wires the card database in
// at boot (NonlegendaryCreatureCardName), as it wires OutsideCollection.
// With nothing wired the object has only its own name.

// NonlegendaryCreatureCardName reports whether `name` is the name of a
// nonlegendary creature card (a face of one, CR 201.4b-d). Wired by the
// server from the Scryfall index (deck.Collection); nil in a bare engine,
// where no name answers yes. A package hook, never a field of Game.
var NonlegendaryCreatureCardName func(name string) bool

// HasAllNonlegendaryCreatureNames reports whether the object has every
// nonlegendary creature card's name in addition to its own: the layer 3
// flag (Characteristic.AllNonlegendaryCreatureNames). A face-down object
// has no name at all (CR 708.2a).
func HasAllNonlegendaryCreatureNames(c Card) bool {
	return !c.FaceDown && c.Effective().AllNonlegendaryCreatureNames
}

// hasNameOfNonlegendaryCreatureCard is the CR 201.2a half the name
// readers add: `c` has all those names and `named` is one of them.
func hasNameOfNonlegendaryCreatureCard(c Card, named string) bool {
	want := strings.TrimSpace(named)
	if want == "" || NonlegendaryCreatureCardName == nil || !HasAllNonlegendaryCreatureNames(c) {
		return false
	}
	return NonlegendaryCreatureCardName(want)
}
