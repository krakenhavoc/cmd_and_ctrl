package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Double Vision — Enchantment {3}{R}{R}:
//
//	"Whenever you cast your first instant or sorcery spell each turn,
//	 copy that spell. You may choose new targets for the copy."
//
// The "first each turn" test is the cast tally Arcane Bombardment reads
// (youCastYourFirstInstantOrSorceryEachTurn), bumped before the cast
// event. The copy is made under your control from last-known
// information if the original has already left the stack (CR 608.2h),
// with the new-targets choice its controller is owed (CR 707.10c). It is
// not a cast, so it does not trigger Double Vision again (CR 707.10).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6629a48b-be9a-4636-8add-69d096a1ca7e",
		Name:         "Double Vision",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast,
				effects.AllOf(effects.YouCast(effects.Or(effects.Instant(), effects.Sorcery())), effects.YouCastYourFirstInstantOrSorceryEachTurn),
				"Double Vision — copy that spell; you may choose new targets for the copy",
				effects.CopyTheTriggeringSpellWithNewTargets),
		},
	})
}
