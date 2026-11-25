package effects

import "fmt"

// cant_spend_mana.go — #6227: the boot check for Spec.CantSpendMana,
// "You can't spend mana to cast this spell" (Hogaak, Arisen
// Necropolis).
//
// The clause takes CR 106.1's one way of paying a mana cost away and
// gives nothing back, so the only way the spell can be cast at all is
// a payment made of something that is not mana. The engine expresses
// two: convoke's taps (CR 702.51a) and delve's exiles (CR 702.66a),
// both of which subtract from the total cost in the one pricer. A card
// declaring the clause with neither could never be cast — which is
// weaker than printed and, since exactly one printed card says it and
// that card has both, a card file's mistake. Refused here rather than
// shipped as an uncastable row.

// CheckCantSpendMana holds "You can't spend mana to cast this spell" to
// the one printed shape: beside convoke or delve, which is what is left
// to pay the mana with.
func CheckCantSpendMana(spec Spec) {
	if !spec.CantSpendMana {
		return
	}
	if spec.TapCost == nil && !spec.Delve {
		panic(fmt.Sprintf("effects.Register: %q can't spend mana to cast itself and has neither a tap cost nor delve — nothing could ever pay for it", spec.Name))
	}
}
