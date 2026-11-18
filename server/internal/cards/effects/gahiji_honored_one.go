package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gahiji, Honored One — Legendary Creature — Beast {2}{R}{G}{W}, 4/4:
//
//	"Whenever a creature attacks one of your opponents or a planeswalker
//	 an opponent controls, that creature gets +2/+0 until end of turn."
//
// Every creature's attack triggers it separately (one trigger per
// attacker), whoever controls the attacker, Gahiji's controller included.
// The attack must be aimed at one of Gahiji's controller's opponents
// themself or at a planeswalker one of them controls (CR 508.3a); an
// attack at Gahiji's controller, their planeswalkers or a battle does
// not trigger. The boost goes to the attacker named by the event, and is
// skipped if it has left the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "110d5d45-7174-4d1b-9e9c-e1ccd950ee3b",
		Name:         "Gahiji, Honored One",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, gahijiAttackAimedAtAnOpponent,
				"Gahiji, Honored One — that creature gets +2/+0 until end of turn",
				gahijiBoostTheAttacker),
		},
	})
}

// gahijiAttackAimedAtAnOpponent is "a creature attacks one of your
// opponents or a planeswalker an opponent controls".
func gahijiAttackAimedAtAnOpponent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack {
		return false
	}
	if opp := attackedPlayer(g, ev); opp != uuid.Nil {
		return opp != source.Controller
	}
	if g.ClassifyAttackTargetForEffect(ev.Target) != game.AttackTargetPlaneswalker {
		return false
	}
	pw, ok := g.LookupCardForEffect(ev.Target)
	return ok && pw.Controller != source.Controller
}

func gahijiBoostTheAttacker(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if !onBattlefield(g, id) {
		return nil
	}
	return BoostUntilEOT{Target: id, Power: 2, Toughness: 0, Label: "Gahiji, Honored One — +2/+0"}.Apply(NewContext(g, item))
}
