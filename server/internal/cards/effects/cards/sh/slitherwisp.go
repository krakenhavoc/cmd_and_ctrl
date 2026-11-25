package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slitherwisp — Creature — Elemental Nightmare {U}{B}{B}, 3/2 (Ikoria: Lair of Behemoths, #5512):
//
//	"Flash
//	 Whenever you cast another spell that has flash, you draw a card and each
//	 opponent loses 1 life."
//
// "Has flash" is the keyword on the spell as it is cast, so an instant, which
// only may be cast any time, does not count; flash granted to a spell does
// (CR 702.36). Slitherwisp's own cast never triggers it, because it is not on
// the battlefield to see it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ccc4f21f-cf8c-40bb-a6d6-1f624320c5b2",
		Name:            "Slitherwisp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.S82OldB115AnotherSpellWithFlashCastByYou, "Slitherwisp — you draw a card and each opponent loses 1 life",
				effects.Do(effects.DrawCards{N: 1}, effects.EachOpponentLosesLife{N: 1})),
		},
	})
}
