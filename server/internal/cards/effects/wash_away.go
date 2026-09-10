package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wash Away — Instant {U}:
//
//	"Cleave {1}{U}{U} (You may cast this spell for its cleave cost.
//	 If you do, remove the words in square brackets.)
//	 Counter target spell [that wasn't cast from its owner's hand]."
//
// The first card in the catalog whose alternative cost widens a
// target clause rather than deleting it. Hard-cast for {U} it is a
// narrow answer — commanders and anything cast off an impulse-exile
// grant; cleaved for {1}{U}{U} it is a plain Counterspell.
//
// "Wasn't cast from its owner's hand" is enforced, not waved through.
// It needed one new fact on the wire: StackItem.CastFromZone, stamped
// at announce because CR 601.2a moves the card to the stack a moment
// later and nothing on the card remembers where it came from. In this
// engine a spell cast from a hand was cast from its OWNER's hand —
// you can only cast out of your own — so the clause is exact rather
// than approximated.
//
// Sandbox simplification: "cast this turn" is not checked. A spell on
// the stack was cast this turn in every situation the engine can
// produce (the stack empties before a turn ends, and nothing here
// suspends a spell across turns), so the clause is satisfied by
// construction rather than by a timestamp. If a card that parks a
// spell on the stack across turns ever lands, this needs a real
// cast-turn stamp.
func init() {
	Register(Spec{
		OracleID: "a4630da0-fe9b-4ead-9621-eac4b7825c35",
		Name:     "Wash Away",
		Targets: TargetSpell("target spell that wasn't cast from its owner's hand",
			Not(CastFromOwnersHand())),
		AlternativeCosts: []game.AlternativeCost{
			Cleave("{1}{U}{U}", TargetSpell("target spell")),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			// One clause or the other, but the effect is the same
			// either way — cleave removes the restriction, not the
			// counter. The engine already validated the target under
			// whichever clause was in force and re-checked it here
			// (CR 608.2b), so this reads the slot and counters.
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
