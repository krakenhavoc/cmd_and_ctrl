package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Toralf's Disciple — Creature — Human Warrior {2}{R}, 3/3 (Mystery
// Booster 2, #5948):
//
//	"Haste
//	 Whenever Toralf's Disciple attacks, conjure four cards named
//	 Lightning Bolt into your library, then shuffle."
//
// Four real Lightning Bolts (ADR 0099 amendment 2026-11-24), then one
// shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5bce8636-aa2f-4f0f-b374-60c1eb1d360a",
		Name:            "Toralf's Disciple",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Toralf's Disciple — conjure four Lightning Bolts into your library, then shuffle",
				effects.Do(effects.CheckConjure("Toralf's Disciple", effects.Conjure{
					Names:   []string{"Lightning Bolt", "Lightning Bolt", "Lightning Bolt", "Lightning Bolt"},
					Zone:    game.ZoneLibrary,
					Shuffle: true,
				}))),
		},
	})
}
