package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Selvala, Explorer Returned — Legendary Creature — Elf Scout {1}{G}{W},
// 2/4 (Murders at Karlov Manor Commander):
//
//	"Parley — {T}: Each player reveals the top card of their library. For
//	 each nonland card revealed this way, add {G} and you gain 1 life.
//	 Then each player draws a card. (Activate only as an instant.)"
//
// A MANA ability (CR 605.1a: it has no target and could add mana), so it
// resolves at once with no stack and the whole text is its Rider: the
// reveals, the mana and the life for each nonland card, then the draws.
// The mana is added by the rider rather than declared in Produced because
// how much there is depends on what the reveals turn up. Summoning
// sickness applies to the {T} cost as usual.
func init() {
	Register(Spec{
		OracleID:     "b628a39b-c923-473d-900a-1617f47ecb76",
		Name:         "Selvala, Explorer Returned",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, LifeGain: 1},
		ManaAbilities: []ManaAbility{{
			Cost:  ManaAbilityCost{Tap: true},
			Label: selvalaLabel,
			Rider: selvalaParley,
		}},
	})
}

const selvalaLabel = "Parley — each player reveals the top card of their library; for each nonland card revealed this way, add {G} and you gain 1 life; then each player draws a card"

func selvalaParley(g *game.Game, controller, source uuid.UUID) error {
	return cmdr2S49Parley(g, source, selvalaLabel, func(nonland int) error {
		if nonland == 0 {
			return nil
		}
		if err := g.AddManaForEffect(controller, source, strings.Repeat("{G}", nonland)); err != nil {
			return err
		}
		return g.ChangePlayerLifeForEffect(source, controller, nonland)
	})
}
