package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blitz of the Thunder-Raptor — Instant {1}{R}:
//
//	"Blitz of the Thunder-Raptor deals damage to target creature or
//	 planeswalker equal to the number of instant and sorcery cards in
//	 your graveyard. If that creature or planeswalker would die this
//	 turn, exile it instead."
//
// The instant and sorcery cards are counted as the spell resolves
// (CR 608.2h), with the Blitz itself still on the stack. The replacement
// is the spell's (ADR 0108 §1): the target is marked even when the count
// is 0 or the damage is prevented.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2db3f9c-26fe-487b-bbb7-8bc6f33456f1",
		Name:         "Blitz of the Thunder-Raptor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve:    effects.DamageFirstTargetExileIfItDies(effects.G2CardsInYourGraveyard(effects.MatchInstantOrSorcery)),
	})
}
