package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gideon's Phalanx — Instant {5}{W}{W} (Magic Origins, #5549):
//
//	"Create four 2/2 white Knight creature tokens with vigilance.
//	 Spell mastery — If there are two or more instant and/or sorcery cards in
//	 your graveyard, creatures you control gain indestructible until end of
//	 turn."
//
// The four Knights, then the grant to every creature the caster controls as
// the spell resolves (CR 611.2c), the new tokens included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f9489831-a9a2-47ab-abae-2f389c9fae2c",
		Name:         "Gideon's Phalanx",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 4, CreatureTokens: 4, Answers: game.AnswerSurvive},
		OnResolve: effects.ResolveSteps(
			effects.CreateToken{Template: effects.TokenCard("2/2 white Knight with vigilance"), N: 4},
			effects.If{Cond: effects.OldB318SpellMastery(), Then: []effects.Applier{
				effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"indestructible"},
					Label:    "Gideon's Phalanx — indestructible",
				},
			}},
		),
	})
}
