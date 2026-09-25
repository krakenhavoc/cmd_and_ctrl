package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nemesis of Reason — Creature — Leviathan Horror {3}{U}{B}, 3/7
// (EDHREC rank 4541):
//
//	"Whenever this creature attacks, defending player mills ten
//	 cards."
//
// The Dimir mill deck's clock. Ten cards a swing is a third of a
// Commander library, and the 3/7 body is built to survive the combat
// it started — it attacks into a board it cannot profitably be
// blocked by and keeps going. The Leviathan does not have to connect:
// the trigger fires on the ATTACK DECLARATION, so a chump block, a
// removal spell after blockers, and a fog all leave the mill intact.
//
// "Defending player" is not a target — it is read off the attack
// event, so no choice is offered and nothing can be redirected. The
// player is captured when the trigger is built and re-checked for
// still being seated at resolution; an opponent eliminated in
// response mills nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b1d70845-869b-45f7-b4bf-1549bac51d58",
		Name:         "Nemesis of Reason",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			},
			Key: "Nemesis of Reason — defending player mills ten cards",
			// The defending player is a fact about the attack
			// DECLARATION (CR 506.4) — which player's planeswalker or
			// battle the attacker is going after — so it is read once,
			// in Build, off the board as it stood then.
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Nemesis of Reason — defending player mills ten cards", nil)
				item.Params.Player = b17DefendingPlayer(g, ev)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				defender := item.Params.Player
				if defender == uuid.Nil || g.PlayerByIDForEffect(defender) == nil {
					return nil
				}
				return MillCards{Player: defender, N: 10}.Apply(NewContext(g, item))
			},
		}},
	})
}
