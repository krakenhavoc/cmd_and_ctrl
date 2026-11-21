package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Metropolis Angel — Creature — Angel Soldier {2}{W}{U}, 3/1 (Streets of
// New Capenna):
//
//	"Flying
//	 Whenever you attack with one or more creatures with counters on
//	 them, draw a card."
//
// "One or more" is one trigger per attack declaration (OncePerBatch, Anim
// Pakal's shape), and it needs an attacker of yours that carries a counter
// of any kind as it is declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c29b592c-7115-4464-beba-e79f62691593",
		Name:            "Metropolis Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && effects.S82SncHasACounter(c)
			}, "Metropolis Angel — draw a card",
				effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}
