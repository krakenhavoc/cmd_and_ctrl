package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon Whisperer — Creature — Human Shaman {R}{R}, 2/2 (Dragons of Tarkir, S82
// older sets):
//
//	"{R}: This creature gains flying until end of turn.
//	 {1}{R}: This creature gets +1/+0 until end of turn.
//	 Formidable — {4}{R}{R}: Create a 4/4 red Dragon creature token with flying.
//	 Activate only if creatures you control have total power 8 or greater."
//
// Two ordinary self-pump rows (Afterthought Sentry's flying grant, the shared
// ThisGetsRow) and a token ability behind the formidable condition (CR 602.5b),
// read as the ability is announced. The Dragon is the printed 4/4 red flier.
//
// No simplification.
func init() {
	const name = "Dragon Whisperer"
	effects.Register(effects.Spec{
		OracleID:     "54884040-b426-43da-9bb1-016621e1446d",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}: This creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{R}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", name+" — flying until end of turn"),
			},
			effects.ThisGetsRow("{1}{R}", 1, 0, name),
			effects.OldB324FormidableRow(effects.ActivatedAbility{
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{4}{R}{R}"),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 red Dragon with flying"), N: 1}),
			}, "{4}{R}{R}: Create a 4/4 red Dragon creature token with flying."),
		},
	})
}
