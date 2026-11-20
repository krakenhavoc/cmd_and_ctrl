package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Helm of Chatzuk — Artifact {1} (Limited Edition Alpha):
//
//	"{1}, {T}: Target creature gains banding until end of turn."
//
// The grant is an ordinary keyword grant until end of turn, and banding
// is read off the layered ability list everywhere the engine asks for it
// (game/banding.go, ADR 0224): a creature that gained it may form a band
// as attackers are declared, and its controller divides the damage of a
// creature blocking it or blocked by it.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "948e3bb7-8265-4e95-acd1-a0c4f22441df",
		Name:         "Helm of Chatzuk",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Target creature gains banding until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, game.KeywordBanding)),
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  ebTargetCreatureGainsUntilEOT(game.KeywordBanding, "Helm of Chatzuk — banding until end of turn"),
		}},
	})
}
