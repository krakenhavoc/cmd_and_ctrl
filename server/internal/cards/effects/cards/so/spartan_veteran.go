package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spartan Veteran — Creature — Human Soldier {R}, 1/1 (Assassin's
// Creed, #5466):
//
//	"During your turn, this creature has first strike.
//	 {2}: This creature gets +1/+0 until end of turn."
//
// Inspiring Paladin's turn-gated first strike on itself, and Fire Nation
// Cadets' firebreathing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b255d1ec-49ac-49a2-9af8-21915bbcd224",
		Name:         "Spartan Veteran",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.IsActivePlayer(g, source.Controller)
			}, "first strike"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Spartan Veteran — +1/+0 until end of turn"),
		}},
	})
}
