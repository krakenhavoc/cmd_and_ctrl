package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inside Source — Creature — Human Citizen {2}{W}, 1/1 (Murders at Karlov
// Manor, #4469):
//
//	"When this creature enters, create a 2/2 white and blue Detective
//	 creature token.
//	 {3}, {T}: Target Detective you control gets +2/+0 and gains vigilance
//	 until end of turn. Activate only as a sorcery."
//
// The Detective is Museum Nightwatch's token. The ability's body is
// Might of the Ancestors' pump (pumpAndGrantFirstTarget), over a target
// that must be a Detective its controller controls, judged at activation
// and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Inside Source — +2/+0 and vigilance until end of turn"
	Register(Spec{
		OracleID:     "12bced67-69bc-4fc7-a585-577060afdd50",
		Name:         "Inside Source",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Inside Source — create a 2/2 white and blue Detective creature token",
				Do(CreateToken{Template: TokenCard("2/2 white and blue Detective"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "{3}, {T}: Target Detective you control gets +2/+0 and gains vigilance until end of turn. " +
				"Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}"), TapCost()),
			Targets:      TargetCreature("target Detective you control", HasSubtype("Detective"), YouControl()),
			SorcerySpeed: true,
			Purpose:      ForTargets(BenefitUntilEOT(0, 2, 0, "vigilance")),
			Effect:       pumpAndGrantFirstTarget(2, 0, label, "vigilance"),
		}},
	})
}
