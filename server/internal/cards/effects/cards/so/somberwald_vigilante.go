package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Somberwald Vigilante — Creature — Human Warrior {R}, 1/1 (Duel Decks:
// Heroes vs. Monsters, S82 old sets):
//
//	"Whenever this creature becomes blocked by a creature, this creature deals
//	 1 damage to that creature."
//
// Kolaghan Aspirant's text, through one shared body: a trigger per blocker
// (CR 509.3d), and the damage goes to that blocker only if it is still on the
// battlefield as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "31c77036-8627-476a-996b-2a6b005c9e12",
		Name:         "Somberwald Vigilante",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.ViciousBattleragerBlocked,
				"Somberwald Vigilante — it deals 1 damage to the creature blocking it",
				effects.S82OldB398PingTheBlocker(1)),
		},
	})
}
