package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Minotaur, Roxxon CEO — Legendary Creature — Minotaur Villain
// {4}{B}{B}, 4/4 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever Minotaur, Roxxon CEO or another nontoken creature dies,
//	 you create a 2/1 black Villain creature token with menace."
//
// Any player's nontoken creature counts, Minotaur's own death included
// (a leaves-the-battlefield trigger looks back at the creature as it
// last existed, CR 603.10a), and "nontoken" is read off that
// last-known object. The token is the shared Villain-with-menace row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f81207e8-d5a2-4a5e-8a81-803ac563fe76",
		Name:         "Minotaur, Roxxon CEO",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && !effects.IsToken(dead)
			}, "Minotaur, Roxxon CEO — create a 2/1 black Villain creature token with menace",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/1 black Villain with menace"), N: 1})),
		},
	})
}
