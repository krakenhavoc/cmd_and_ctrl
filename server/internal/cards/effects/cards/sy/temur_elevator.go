package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Temur Elevator — Land:
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 {T}: Add {G}, {U}, or {R}. If you don't have the city's blessing, you
//	 lose 1 life."
//
// The life loss is the mana ability's rider, run once the mana is in the
// pool and judged then (CR 605.3b): whether its controller has the city's
// blessing is read at that moment, and it is a loss, not damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cc36ec3e-ef7e-4d3c-84f2-80ee13cbad26",
		Name:            "Temur Elevator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G|U|R}",
			Label:    "Add {G}, {U}, or {R}. If you don't have the city's blessing, you lose 1 life",
			Rider: func(g *game.Game, controller, source uuid.UUID) error {
				if effects.YouHaveTheCitysBlessing(g, controller) {
					return nil
				}
				return g.ChangePlayerLifeForEffect(source, controller, -1)
			},
		}},
	})
}
