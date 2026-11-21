package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gruul Keyrune — Artifact {3} (Gatecrash, #5565):
//
//	"{T}: Add {R} or {G}.
//	 {R}{G}: This artifact becomes a 3/2 red and green Beast artifact
//	 creature with trample until end of turn."
//
// Dimir Keyrune's mana ability, and Levitating Statue's animation with
// the colours and trample added (gtcKeyruneAnimate). It is still an
// artifact; a Keyrune that came under your control this turn can animate
// but can't attack or tap for mana while it is a creature (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98bdb1e1-7550-47e8-aa06-f81802b257a0",
		Name:         "Gruul Keyrune",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R|G}",
			Label:    "Add {R} or {G}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}{G}: This artifact becomes a 3/2 red and green Beast artifact creature with trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{R}{G}"),
			Effect: effects.GtcKeyruneAnimate("Gruul Keyrune — a 3/2 red and green Beast artifact creature with trample until end of turn",
				3, 2, "Beast", []string{"R", "G"}, "trample"),
		}},
	})
}
