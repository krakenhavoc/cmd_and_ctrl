package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mercurial Geists — Creature — Spirit {2}{U}{R}, 1/3 (Eldritch Moon,
// #5541):
//
//	"Flying
//	 Whenever you cast an instant or sorcery spell, this creature gets
//	 +3/+0 until end of turn."
//
// Kiln Fiend's cast trigger on a flyer.
//
// No simplification.
func init() {
	const label = "Mercurial Geists — +3/+0 until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "1f6bff6c-a2f0-4ed7-8169-acb4d31a7187",
		Name:            "Mercurial Geists",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()), label, effects.ThisGetsUntilEndOfTurn(3, 0, label)),
		},
	})
}
