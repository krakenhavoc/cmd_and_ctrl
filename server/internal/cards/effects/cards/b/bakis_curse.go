package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Baki's Curse — Sorcery {2}{U}{U} (Homelands, S82 OLD b824, #5635):
//
//	"Baki's Curse deals 2 damage to each creature for each Aura attached to that
//	creature."
//
// One damage instance (CR 120.1) from the spell, each creature taking 2 times the
// Auras attached to it as it resolves (B18AurasAttachedTo, Kor Spiritdancer's
// count); a creature with no Aura takes none. It is a board sweep that only
// touches enchanted creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "341a92b9-d6a0-499e-abb7-e2b7d7bc248c",
		Name:         "Baki's Curse",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
		OnResolve:    effects.OldB824DamageEachCreatureForEachAura(2),
	})
}
