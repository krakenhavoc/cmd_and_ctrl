package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Starting Town — Land — Town (EDHREC rank 743):
//
//	"This land enters tapped unless it's your first, second, or third
//	 turn of the game.
//	 {T}: Add {C}.
//	 {T}, Pay 1 life: Add one mana of any color."
//
// An untapped five-colour land in the opening turns, a painful one
// later. The colourless half is free and sits first so the auto-tapper
// reaches for it; the coloured half pays a life as a real COST (Mana
// Confluence's shape — refused at 0 life without tapping), and prints
// "any color", so no commander-identity narrowing.
func init() {
	Register(Spec{
		OracleID:     "d04e0975-f401-41b8-a9db-9bcf9cbbce66",
		Name:         "Starting Town",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(func(g *game.Game, controller uuid.UUID) bool {
			return g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) &&
				g.Seats[g.Turn.ActiveSeat].ID == controller && g.TurnsBegunFor(controller) <= 3
		})},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Life: 1},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}, Pay 1 life: Add one mana of any color",
			},
		},
	})
}
