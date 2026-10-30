package effects

// Scientist Supreme of A.I.M. — Legendary Creature — Human Scientist Villain
// {U}{B}, 2/2 (Marvel Super Heroes, #4455):
//
//	"Pay 2 life: Copy target activated or triggered ability you control from
//	 an artifact source. You may choose new targets for the copy. Activate
//	 only during your turn and only once each turn."
//
// Echo, Perceptive Prodigy's ability copy with an artifact-source filter. A
// mana ability is not on the stack and cannot be targeted. The two
// restrictions are announce-time conditions (CR 602.5b), the once-per-turn
// one counted per object and per ability so a second Scientist has its own.
//
// No simplification.
func init() {
	const label = "Pay 2 life: Copy target activated or triggered ability you control from an artifact source. You may choose new targets for the copy. Activate only during your turn and only once each turn."
	Register(Spec{
		OracleID:     "c09c424b-6aba-4190-9f55-f0307d746386",
		Name:         "Scientist Supreme of A.I.M.",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: label,
			Cost:  PayLife(2),
			Targets: AbilityOnStack("target activated or triggered ability you control from an artifact source",
				AnAbilityYouControl(), AbilityFromSource(Artifact())),
			Condition: AllConditions(DuringYourTurn(), OncePerTurnActivation(label)),
			Effect:    copyTargetedAbility,
		}},
	})
}
