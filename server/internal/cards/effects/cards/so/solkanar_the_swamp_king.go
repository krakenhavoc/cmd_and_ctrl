package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sol'kanar the Swamp King — Legendary Creature — Demon {2}{U}{B}{R}, 5/5
// (Dominaria Remastered):
//
//	"Swampwalk
//	 Whenever a player casts a black spell, you gain 1 life."
//
// Demon's Horn's trigger without the "may": any player's cast, the
// Demon's controller's included (TmcB8APlayerCast), the colour read
// through the Demon's text. It resolves even if the spell is countered
// in response. Swampwalk is the engine's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b8aac8e8-4588-4cb9-8b95-5c4abab1c48b",
		Name:            "Sol'kanar the Swamp King",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.TmcB8APlayerCast(effects.OfColor("B")),
				"Sol'kanar the Swamp King — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
