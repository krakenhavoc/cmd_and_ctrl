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
//
// Sandbox simplification: "your first, second, or third turn" is read
// as the game's first three ROUNDS (Turn.Number, which counts rounds
// and only advances when the table wraps). Those are the same thing
// until someone takes an extra turn, and no catalog card grants one.
func init() {
	Register(Spec{
		OracleID:     "d04e0975-f401-41b8-a9db-9bcf9cbbce66",
		Name:         "Starting Town",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Your first three turns are counted as the game's first three rounds, which only differs if someone takes an extra turn."},
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(func(g *game.Game, _ uuid.UUID) bool {
			return g.Turn.Number <= 3
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
