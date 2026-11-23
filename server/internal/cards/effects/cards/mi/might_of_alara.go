package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Might of Alara — Instant {G} (Conflux, #5587):
//
//	"Domain — Target creature gets +1/+1 until end of turn for each
//	 basic land type among lands you control."
//
// Gaea's Might's text. The count is the shared DomainCount (domain.go,
// CR 305.6), read once as the spell resolves (CR 608.2h), so a land
// played afterwards changes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "44570849-b2b0-47a8-9d77-f7c738896a0a",
		Name:         "Might of Alara",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit}),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.ConB1DomainBoostTheTarget(1, "Might of Alara — +1/+1 for each basic land type"),
	})
}
