package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Terrifying Presence — Instant {1}{G}:
//
//	"Prevent all combat damage that would be dealt by creatures other
//	 than target creature this turn."
//
// #2026: the target is pinned as the object it is as the spell resolves
// (CR 400.7) and left out of the shield; every other creature's combat
// damage is prevented, read as it would be dealt (CR 609.7b), so a
// creature that enters later is caught. The ruling: a target that has
// become illegal stops the spell, and nothing is prevented (CR 608.2b).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca8f6e19-6ef3-4d77-8312-867767ffeda6",
		Name:         "Terrifying Presence",
		Purpose:      game.Purpose{Answers: game.AnswerPrevent}, // #4734 slice 6
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.SourceShieldSpell(effects.CombatShieldAgainstCreatures(game.DamageSourceFilter{}).OtherThanTarget(0)),
	})
}
