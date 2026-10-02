package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rampaging Ferocidon — Creature — Dinosaur {2}{R}, 3/3:
//
//	"Menace
//	 Players can't gain life.
//	 Whenever another creature enters, this creature deals 1 damage to
//	 that creature's controller."
//
// Menace rides PrintedKeywords; "Players can't gain life" is ADR 0107
// §5's battlefield static (CR 119.7, #1880), its controller included.
//
// The trigger watches every player's creatures, the controller's own
// included ("another creature", not "another creature an opponent
// controls"). "That creature's controller" is read off the entering
// object's last-known information, captured when the ability triggered
// (the trigger context's snapshot), so a creature that leaves or
// changes hands in response still pings the player who controlled it
// as it entered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3e5ca524-8dd1-4f7f-a467-5210d768b8a1",
		Name:            "Rampaging Ferocidon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		CantGainLife:    PlayersCantGainLife(),
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == uuid.Nil || ev.CardID == source.InstanceID {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature()
			}, "Rampaging Ferocidon — 1 damage to that creature's controller", rampagingFerocidonPing),
		},
	})
}

// rampagingFerocidonPing deals 1 damage to the entering creature's
// controller as it entered.
func rampagingFerocidonPing(g *game.Game, item *game.StackItem) error {
	obj := NewContext(g, item).Trigger().Object
	if obj == nil || obj.Controller == uuid.Nil {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: obj.Controller, Amount: 1}.Apply(NewContext(g, item))
}
