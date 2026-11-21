package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Konstrari Improviser // Soul Tether — Creature — Human Artificer
// {1}{R/G}, 2/2 // Sorcery {2}{R/G} (preparation card, CR 722):
//
//	"This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)"
//
//	Soul Tether — "Create a Heartwood token. (It's a red and green
//	 artifact with '{T}: Add {R} or {G}.')"
//
// Soul Tether is shared with Heartwood Crafter (soulTetherResolve).
//
// No simplification.
func init() {
	const id = "bad2e3b2-ad47-4e17-b35e-0fcc13e82937"
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Konstrari Improviser",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersPrepared()},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Soul Tether",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.SoulTetherResolve,
	})
}
