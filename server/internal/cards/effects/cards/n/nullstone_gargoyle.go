package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nullstone Gargoyle — Artifact Creature — Gargoyle {9}, 4/5 (Ravnica: City of
// Guilds, S82 OLD b582):
//
//	"Flying
//	 Whenever the first noncreature spell of a turn is cast, counter that
//	 spell."
//
// Jace's "counter that spell" on the cast trigger, with Esper Sentinel's tally
// read across every player's casts (the tallies count a spell cast before the
// Gargoyle arrived, and are bumped before EventCast fires). Any player's
// spell, the Gargoyle's controller's included. A spell that can't be
// countered stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7b175a33-a05c-49fa-802a-36f7013c892a",
		Name:            "Nullstone Gargoyle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.S82OldB582FirstNoncreatureSpellOfTheTurn,
				"Nullstone Gargoyle — counter that spell", effects.S82OldB582CounterThatSpell),
		},
	})
}
