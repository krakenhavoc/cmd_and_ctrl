package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rodeo Pyromancers — Creature — Human Mercenary {3}{R}, 3/4 (Outlaws of
// Thunder Junction, #4468):
//
//	"Whenever you cast your first spell each turn, add {R}{R}."
//
// Rashmi, Eternities Crafter's first-spell clock (rashmiFirstSpell: the
// per-turn cast tally is bumped before EventCast, so a total of one is
// the first) on every turn, not only yours. The ability triggers on a
// cast, not on a mana ability, so it is no mana ability (CR 605.1b): it
// uses the stack, and the {R}{R} is AddMana's, which empties with the
// pool at the end of the step (CR 106.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82debe1c-31cb-411f-a353-98a952180666",
		Name:         "Rodeo Pyromancers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.RashmiFirstSpell, "Rodeo Pyromancers — add {R}{R}", effects.Do(effects.AddMana{Produced: "{R}{R}"})),
		},
	})
}
