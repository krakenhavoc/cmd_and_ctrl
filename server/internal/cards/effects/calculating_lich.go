package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Calculating Lich — Creature — Zombie Wizard {4}{B}{B}, 5/5 (Game Night
// 2019, #5515):
//
//	"Menace
//	 Whenever a creature attacks one of your opponents, that player loses 1
//	 life."
//
// "A creature attacks one of your opponents" is M'Baku's reading (CR
// 508.3a): the creature was declared attacking that opponent ITSELF
// (attackedPlayer), so an attack on an opponent's planeswalker or battle
// does not trigger it, and an attack on the Lich's controller is not one
// either. The creature may be anyone's, so a fight between two opponents
// still bleeds the defender. It triggers once per attacking creature, and
// only a DECLARED attacker triggers it (EventAttack): a creature
// put onto the battlefield attacking does not.
//
// The life loss reads the attacked player back off the trigger's event as
// it resolves, so it lands on the defender even if the attacker has left
// combat; a defender who has left the game loses nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7264d3b3-fd46-4ea3-a85d-f0b068c331ad",
		Name:            "Calculating Lich",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				opp := attackedPlayer(g, ev)
				return opp != uuid.Nil && opp != source.Controller
			}, "Calculating Lich — the attacked player loses 1 life", calculatingLichDrain),
		},
	})
}

// calculatingLichDrain is "that player loses 1 life": the player the
// triggering creature was declared attacking.
func calculatingLichDrain(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	victim := attackedPlayer(g, item.Trigger.Event)
	if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
		return nil
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -1)
}
