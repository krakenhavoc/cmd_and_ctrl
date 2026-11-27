package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Viridian Joiner — Creature — Elf Druid {2}{G}, 1/2 (Mirrodin, S82 OLD,
// #5606):
//
//	"{T}: Add an amount of {G} equal to this creature's power."
//
// A mana ability (CR 605.1a) whose amount is read as it is activated, from
// the Joiner's power through the layers (Bighorner Rancher's amount with
// the Joiner's own power): counters, a pump or a shrink change what it
// makes, and a power of 0 or less makes nothing. It can't be used the turn
// the Joiner arrives (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84244aa6-11be-4940-b102-e029f955fc9a",
		Name:         "Viridian Joiner",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			Label:          "Add an amount of {G} equal to this creature's power.",
			ProducedAmount: viridianJoinerGreen,
		}},
	})
}

func viridianJoinerGreen(g *game.Game, _, source uuid.UUID) game.ManaAmount {
	g.RecomputeLayersIfStaleLocked()
	n := 0
	if c, ok := g.LookupCardForEffect(source); ok {
		n = c.CurrentPower()
	}
	return game.ManaAmount{Options: []string{"G"}, N: n}
}
