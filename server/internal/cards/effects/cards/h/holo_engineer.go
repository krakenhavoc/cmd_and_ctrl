package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Holo-Engineer — Creature — Human Artificer {2}{G}{G}, 3/3 (Star Trek,
// #4973):
//
//	"When this creature enters, create a 3/3 colorless Illusion artifact
//	 creature token.
//	 If one or more tokens would be created under your control, that many
//	 3/3 colorless Illusion artifact creature tokens are created instead."
//
// Divine Visitation's CR 614.1a replacement over EVERY token kind, not
// only creatures: each group of tokens an instruction would create under
// your control becomes that many Illusions, so a Treasure, a Clue and a
// copy token are Illusions too. The Illusion is the printed token (ttrk
// 24). Its own enters trigger makes an Illusion either way.
//
// No simplification.
func init() {
	anyToken := func(game.Card) bool { return true }
	effects.Register(effects.Spec{
		OracleID:     "0d830d0c-3339-49cc-8527-4f4edf787925",
		Name:         "Holo-Engineer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Holo-Engineer — create a 3/3 colorless Illusion artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 colorless Illusion artifact"), N: 1})),
		},
		Replacements: []game.ReplacementEffect{
			effects.HobA9InsteadCreateTokens("Holo-Engineer: 3/3 colorless Illusion artifact creatures instead",
				anyToken,
				func() []game.Card { return []game.Card{effects.TokenCard("3/3 colorless Illusion artifact")} }),
		},
	})
}
