package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "c29b592c-7115-4464-beba-e79f62691593",
		Name:            "Metropolis Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !attackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && s82SncHasACounter(c)
			}, "Metropolis Angel — draw a card",
				Do(DrawCards{N: 1}))),
		},
	})
}
