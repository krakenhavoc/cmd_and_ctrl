package effects

const wirewoodSymbioteLabel = "Return an Elf you control to its owner's hand: Untap target creature."

// Wirewood Symbiote — Creature — Insect {G}, 1/1:
//
//	"Return an Elf you control to its owner's hand: Untap target
//	 creature. Activate only once each turn."
//
// Quirion Ranger's other half, and the second card on #1213's
// return-to-hand seam row. Same two seams, same reading of both: a
// return-to-hand COST paid at announce, and a once-each-turn gate
// read off the ACTIVATION tally rather than the resolution one.
//
// The one thing that differs from the Ranger is who can pay. The
// Symbiote is an INSECT, so she can never return herself — the clause
// says Elf and the filter is the clause. Returning an Elf that has
// already been used for value (a Priest of Titania that tapped for
// mana, a Wood Elves that already fetched) is the line, and the Elf
// coming back means the ETB is available again next turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "67a52a74-9474-4f4e-8785-6ff54078a8ca",
		Name:         "Wirewood Symbiote",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     wirewoodSymbioteLabel,
			Cost:      ReturnAPermanentToHand("an Elf you control", HasSubtype("Elf")),
			Condition: OncePerTurnActivation(wirewoodSymbioteLabel),
			Targets:   TargetCreature("target creature"),
			Effect:    untapTheTarget,
		}},
	})
}
