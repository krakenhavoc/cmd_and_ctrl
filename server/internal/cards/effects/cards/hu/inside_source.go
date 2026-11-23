package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "12bced67-69bc-4fc7-a585-577060afdd50",
		Name:         "Inside Source",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Inside Source — create a 2/2 white and blue Detective creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 white and blue Detective"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{3}, {T}: Target Detective you control gets +2/+0 and gains vigilance until end of turn. " +
				"Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets:      effects.TargetCreature("target Detective you control", effects.HasSubtype("Detective"), effects.YouControl()),
			SorcerySpeed: true,
			Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "vigilance")),
			Effect:       effects.PumpAndGrantFirstTarget(2, 0, label, "vigilance"),
		}},
	})
}
