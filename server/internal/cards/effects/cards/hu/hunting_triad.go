package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hunting Triad — Kindred Sorcery — Elf {3}{G} (Morningtide, #5592):
//
//	"Create three 1/1 green Elf Warrior creature tokens.
//	 Reinforce 3—{3}{G} ({3}{G}, Discard this card: Put three +1/+1
//	 counters on target creature.)"
//
// Three tokens from the shared table, and the shared hand-only reinforce
// row (morReinforce), which, being an activated ability, can be used at
// instant speed even though the card is a sorcery (CR 602.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66bf0c4b-0c36-4d63-a929-5a7427a07279",
		Name:         "Hunting Triad",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("1/1 green Elf Warrior"), N: 3}),
		Activated: []effects.ActivatedAbility{
			effects.MorReinforce(3, "{3}{G}", "Reinforce 3—{3}{G} ({3}{G}, Discard this card: Put three +1/+1 counters on target creature.)"),
		},
	})
}
