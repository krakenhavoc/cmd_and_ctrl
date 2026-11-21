package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mogg Alarm — Sorcery {1}{R}{R} (Nemesis, #5618):
//
//	"You may sacrifice two Mountains rather than pay this spell's mana
//	 cost.
//	 Create two 1/1 red Goblin creature tokens."
//
// Fireblast's alternative cost (SacrificeInstead, CR 118.9): the two
// Mountains are sacrificed as a cost with the spell already on the
// stack, so a countered Mogg Alarm still costs them. The tokens are
// Krenko's Command's shared red Goblin row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c2117d5-4600-489c-be5b-dc0ed69b30ca",
		Name:         "Mogg Alarm",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		AlternativeCosts: []game.AlternativeCost{
			effects.SacrificeInstead(2, "two Mountains", 0, effects.HasSubtype("Mountain")),
		},
		OnResolve: effects.ResolveSteps(effects.CreateToken{Template: effects.RedGoblinToken(), N: 2}),
	})
}
