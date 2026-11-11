package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Selvala, Explorer Returned — Legendary Creature — Elf Scout {1}{G}{W},
// 2/4 (Murders at Karlov Manor Commander):
//
//	"Parley — {T}: Each player reveals the top card of their library. For
//	 each nonland card revealed this way, add {G} and you gain 1 life.
//	 Then each player draws a card. (Activate only as an instant.)"
//
// An ordinary activated ability, not a mana ability: CR 605.1a says an
// ability whose cost or effect moves a card to or from a library is not
// one, and "Then each player draws a card" does (Millikin's rule, #3971;
// the printed "Activate only as an instant" reminder says the same). So it
// uses the stack (CR 602.2), an opponent may respond, and the reveals, the
// mana and the life for each nonland card, then the draws all happen on
// resolution. Summoning sickness applies to the {T} cost as usual.
func init() {
	Register(Spec{
		OracleID:     "b628a39b-c923-473d-900a-1617f47ecb76",
		Name:         "Selvala, Explorer Returned",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   selvalaLabel,
			Cost:    TapCost(),
			Purpose: game.Purpose{Draws: 1, LifeGain: 1, Answers: game.AnswerValue}, // ADR 0142: mana, life and a draw answer nothing
			Effect:  selvalaParley,
		}},
	})
}

const selvalaLabel = "Parley — {T}: Each player reveals the top card of their library. For each nonland card revealed this way, add {G} and you gain 1 life. Then each player draws a card."

func selvalaParley(g *game.Game, item *game.StackItem) error {
	return cmdr2S49Parley(g, item.SourceCardID, selvalaLabel, func(nonland int) error {
		if nonland == 0 {
			return nil
		}
		if err := g.AddManaForEffect(item.Controller, item.SourceCardID, strings.Repeat("{G}", nonland)); err != nil {
			return err
		}
		return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, nonland)
	})
}
