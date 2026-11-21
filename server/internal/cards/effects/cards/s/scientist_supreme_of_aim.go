package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c09c424b-6aba-4190-9f55-f0307d746386",
		Name:         "Scientist Supreme of A.I.M.",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: label,
			Cost:  effects.PayLife(2),
			Targets: effects.AbilityOnStack("target activated or triggered ability you control from an artifact source",
				effects.AnAbilityYouControl(), effects.AbilityFromSource(effects.Artifact())),
			Condition: effects.AllConditions(effects.DuringYourTurn(), effects.OncePerTurnActivation(label)),
			Effect:    effects.CopyTargetedAbility,
		}},
	})
}
