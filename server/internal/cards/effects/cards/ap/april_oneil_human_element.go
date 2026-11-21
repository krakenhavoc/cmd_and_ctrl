package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// April O'Neil, Human Element — Legendary Creature — Human Detective
// {3}{U}, 2/5 (TMNT Commander, #5011):
//
//	"Whenever a player casts an artifact, instant, or sorcery spell, you
//	 create a Mutagen token."
//
// Any player's cast, April's controller's included. The trigger resolves
// above the spell, so the Mutagen arrives even if the spell is countered
// (the card's ruling).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "020963ef-24c0-48cc-8776-bc257df684bc",
		Name:         "April O'Neil, Human Element",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.TmcB7APlayerCasts(effects.TmcB7ArtifactInstantOrSorcery()),
				"April O'Neil, Human Element — create a Mutagen token",
				effects.Do(effects.CreateToken{Template: effects.MutagenToken(), N: 1})),
		},
	})
}
