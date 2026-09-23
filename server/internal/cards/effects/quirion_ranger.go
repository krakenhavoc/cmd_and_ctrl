package effects

const quirionRangerLabel = "Return a Forest you control to its owner's hand: Untap target creature."

// Quirion Ranger — Creature — Elf Ranger {G}, 1/1:
//
//	"Return a Forest you control to its owner's hand: Untap target
//	 creature. Activate only once each turn."
//
// Two seams closing at once, which is why she waited so long: the
// return-to-hand COST (#1213) and the per-source activations-this-turn
// count (#1181's Game.Activations, which had no reader until now).
// Without either she would have shipped stronger than printed, and
// #259 says that is the one direction we do not ship in.
//
// The once-each-turn gate is an ACTIVATION count, not a resolution
// count, and the difference is the whole point on a card like this:
// hold two activations on the stack and a resolution count would let
// you announce the second. It reads the tally per OBJECT (CR 400.7),
// so a Ranger that died and came back is a new object with a fresh
// use — as in paper.
//
// The free untap is the card. Returning a Forest costs nothing but
// the land drop you are about to re-take, and the untap is on an
// ability with no mana cost and no {T}, so she untaps a mana creature
// — or herself, if something has tapped her — every turn for free.
// Paired with Wirewood Symbiote it is the same trick from the other
// side, which is why the two are one seam row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ecaefc8-ead2-47a3-a7ea-b030faab65a7",
		Name:         "Quirion Ranger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     quirionRangerLabel,
			Cost:      ReturnAPermanentToHand("a Forest you control", HasSubtype("Forest")),
			Condition: OncePerTurnActivation(quirionRangerLabel),
			Targets:   TargetCreature("target creature"),
			Effect:    untapTheTarget,
		}},
	})
}
