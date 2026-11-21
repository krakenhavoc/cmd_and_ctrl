package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tzaangor Shaman — Creature — Mutant Shaman {2}{U}{R}, 3/3 (Warhammer 40,000
// Commander, S82 GAP batch 195, #5948):
//
//	"Flying
//	 Sorcerous Elixir — Whenever this creature deals combat damage to a player,
//	 copy the next instant or sorcery spell you cast this turn when you cast it.
//	 You may choose new targets for the copy."
//
// Doublecast's delayed trigger (WhenYouNextCast over an instant or sorcery
// spell, the copy made with the new-targets offer from last-known information),
// scheduled when the Shaman connects. It fires once, for the first instant or
// sorcery you cast this turn after that; a creature spell cast first neither
// fires it nor uses it up, and it ends with the turn. Two combat damage events
// are two promises.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1d55a1c5-fd0b-44d2-9e7c-550052fef262",
		Name:            "Tzaangor Shaman",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Tzaangor Shaman — copy the next instant or sorcery spell you cast this turn when you cast it",
				effects.GapB195CopyTheNextInstantOrSorcery("Tzaangor Shaman — copy that spell")),
		},
	})
}
