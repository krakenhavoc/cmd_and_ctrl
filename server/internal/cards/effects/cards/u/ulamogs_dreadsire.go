package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ulamog's Dreadsire — Creature — Eldrazi {10}, 10/10 (Modern Horizons 3
// Commander, #5468):
//
//	"Vigilance
//	 Ward—Sacrifice a permanent with mana value 1 or greater.
//	 {T}: Create a 10/10 colorless Eldrazi creature token."
//
// Ward is the sacrifice ward (Vein Ripper's): the player whose spell or
// ability targets the Dreadsire sacrifices a permanent they control with
// mana value 1 or greater, or it is countered. Lands, and tokens with no
// mana cost, have mana value 0 and cannot pay it, and a player with no
// such permanent cannot pay at all. The {T} ability is an ordinary
// creature tap ability, so summoning sickness stops it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2afefaab-c614-4fca-afba-d8d206d9a8b9",
		Name:            "Ulamog's Dreadsire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardSacrifice("a permanent with mana value 1 or greater", effects.Permanent(), effects.ManaValueGE(1)),
				"Ulamog's Dreadsire — ward, sacrifice a permanent with mana value 1 or greater"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Create a 10/10 colorless Eldrazi creature token.",
			Cost:    effects.TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("10/10 colorless Eldrazi"), N: 1}),
		}},
	})
}
