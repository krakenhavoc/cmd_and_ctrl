package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archfiend of Depravity — Creature — Demon {3}{B}{B}, 5/4 (EDHREC
// rank 1572):
//
//	"Flying
//	 At the beginning of each opponent's end step, that player chooses
//	 up to two creatures they control, then sacrifices the rest."
//
// The Demon that lets everyone keep two. The trigger fires on every
// opponent's end step (the event's Actor is the active player), and
// the choice is THAT PLAYER's: it is asked through the engine's
// sacrifice prompt, which offers a player their own permanents and
// nobody else's — so hexproof and indestructible are irrelevant, as
// they are on paper, because nothing is targeted or destroyed.
//
// Sandbox simplification, declared: the prompt picks one creature at
// a time, so "choose two to keep" is asked as (creatures − 2)
// prompts to sacrifice one each — the same decision from the other
// side, made by the same player, with the option list trimmed after
// every answer. The prompts are queued together when the ability
// resolves, so in the one corner where a creature leaves between two
// answers for some OTHER reason (a lord sacrificed first, shrinking a
// damaged creature to zero toughness) the remaining prompts still
// ask for the count computed at resolution, and the player can end
// with fewer than two. Nothing can respond between the answers — a
// pending prompt holds priority — so that corner is the only one.
func init() {
	Register(Spec{
		OracleID:        "af247e2f-b271-4f5b-ab98-4579d2c17c21",
		Name:            "Archfiend of Depravity",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The opponent is asked to sacrifice one creature at a time until two remain, rather than choosing the two to keep up front."},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginEndStep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b14OpponentsEndStepBegan(ev, source)
			},
			Key: "Archfiend of Depravity — that player keeps up to two creatures and sacrifices the rest",
			Effect: func(g *game.Game, item *game.StackItem) error {
				var victim uuid.UUID
				if item.Trigger != nil {
					victim = item.Trigger.Event.Actor
				}
				b14PlayerSacrificesAllButN(g, item.SourceCardID, victim, 2, "Archfiend of Depravity — sacrifice a creature (keep up to two)")
				return nil
			},
		}},
	})
}
