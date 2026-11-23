package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grim Draugr — Snow Creature — Zombie Berserker {2}{B}, 3/2 (Kaldheim,
// #5986):
//
//	"{1}{S}: This creature gets +1/+0 and gains menace until end of turn.
//	 (It can't be blocked except by two or more creatures. {S} can be
//	 paid with one mana from a snow source.)"
//
// Fangblade Brigand's thisCreatureUntilEOT body with menace for first
// strike. The {1} is any mana and the {S} one mana from a snow source
// (CR 107.4h, #5986); a generic reduction would reach only the {1}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "793fe164-6c57-4fe0-a955-d5f7ffe388a9",
		Name:         "Grim Draugr",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{S}: This creature gets +1/+0 and gains menace until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}{S}"),
			Effect:  effects.ThisCreatureUntilEOT("Grim Draugr — +1/+0 and menace", 1, 0, "menace"),
		}},
	})
}
