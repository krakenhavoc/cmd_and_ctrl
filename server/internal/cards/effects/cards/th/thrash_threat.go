package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thrash // Threat — split card {R/G}{R/G} // {2}{R}{G} (Ravnica Remastered):
//
//	Thrash (Instant): "Target creature you control deals damage equal to its
//	 power to target creature or planeswalker you don't control."
//	Threat (Sorcery): "Create a 4/4 red and green Beast creature token with
//	 trample."
//
// Thrash is Bite Down's one-sided bite (bitingSpell); Threat is the
// red-and-green Beast row already in the token table.
//
// No simplification.
func init() {
	effects.Register(effects.BitingSpell("342d20ca-178d-433e-9630-6fed33a8d0bc", "Thrash"))
	effects.Register(effects.Spec{
		OracleID:     "342d20ca-178d-433e-9630-6fed33a8d0bc#1",
		Name:         "Threat",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("4/4 red and green Beast with trample"), N: 1}),
	})
}
