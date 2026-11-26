package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Echoing Boon — Conspiracy (Conspiracy: Take the Crown, S82 OLD batch 267,
// #5540):
//
//	"Hidden agenda
//	 Whenever you cast an instant or sorcery spell, if it targets a creature
//	 you control with the chosen name, you may copy that spell and may choose
//	 new targets for the copy."
//
// Hidden agenda (ADR 0227): face down until its owner turns it up, and it
// does nothing before then (CR 315.5). The intervening if is read as the
// spell is cast and again as the ability resolves (CR 603.4,
// S82OldB267CopyTheSpellIfItStillTargetsANamedCreature), so a spell that was
// countered in response is not copied. The copy is created, not cast, so it
// triggers no cast trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22e89c6d-2304-4def-8a74-e53a6271e9b9",
		Name:         "Echoing Boon",
		Completeness: effects.CompletenessFull,
		HiddenAgenda: effects.HiddenAgendaOneName,
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.On(game.EventCast, effects.S82OldB267CastInstantOrSorceryTargetingYourNamedCreature,
					"Echoing Boon — copy that spell; you may choose new targets for the copy",
					effects.S82OldB267CopyTheSpellIfItStillTargetsANamedCreature),
				"Echoing Boon — copy that spell?"),
		},
	})
}
