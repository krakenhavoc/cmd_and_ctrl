package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// single_source_lethal.go — the card side of #6209: "This creature can't
// be destroyed by lethal damage unless lethal damage dealt by a single
// source is marked on it" (Ogre Enforcer).
//
// CR 704.5g destroys a creature whose MARKED damage totals its
// toughness, whatever marked it. This clause asks the one question the
// total cannot answer, so the engine keeps the marked damage per source
// (game/single_source_lethal.go) and the state-based action reads it for
// a creature with this static.
//
//	Spec{
//	    SingleSourceLethalOnly: SingleSourceLethalOnly(
//	        "This creature can't be destroyed by lethal damage unless lethal damage dealt by a single source is marked on it."),
//	}
//
// It is the creature's own static, read under its CatalogAbilityKey, so
// a creature that loses all its abilities is destroyed by lethal damage
// like any other (CR 613.1f). It says nothing about the DEATHTOUCH
// state-based action (CR 704.5h), which destroys for any nonzero damage
// from a deathtouch source and is a rule of its own, nor about a
// "destroy" effect, which is not lethal damage at all.

// SingleSourceLethalOnly is the static, for Spec.SingleSourceLethalOnly.
// `label` is the clause as printed.
func SingleSourceLethalOnly(label string) []game.SingleSourceLethalOnly {
	return []game.SingleSourceLethalOnly{{Label: label}}
}
