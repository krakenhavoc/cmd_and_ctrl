package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rampaging Raptor — Creature — Dinosaur {2}{R}{R}, 4/4 (March of the
// Machine):
//
//	"Trample, haste
//	 {2}{R}: This creature gets +2/+0 until end of turn.
//	 Whenever this creature deals combat damage to an opponent, it deals
//	 that much damage to target planeswalker that player controls or
//	 battle that player protects."
//
// "That player" is the opponent the damage event names; the target is
// chosen from their planeswalkers and the battles they protect (CR 310.9)
// as the trigger goes on the stack, and "that much" is the event's amount,
// read as the trigger resolves. The Raptor deals the damage (as it last
// existed if it has left, CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "450a07f5-709c-48a6-aaa8-ddbd7da4fd2e",
		Name:            "Rampaging Raptor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}{R}"),
			Effect:  thisGetsUntilEndOfTurn(2, 0, "Rampaging Raptor — +2/+0"),
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && ev.Target != source.Controller &&
					combatDamageToPlayerBy(ev, source.Controller, g)
			},
			TargetsFrom: s82MomB12RaptorTargets,
			Key:         "Rampaging Raptor — it deals that much damage to target planeswalker that player controls or battle that player protects",
			Effect:      b24DamageToTargetEqualToEvent,
		}},
	})
}

// s82MomB12RaptorTargets is "target planeswalker that player controls or
// battle that player protects", "that player" being the player the damage
// event names.
func s82MomB12RaptorTargets(tc game.TriggerContext, _ *game.Card, g *game.Game) *game.TargetSpec {
	player := tc.Event.Target
	if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
		return nil
	}
	return &game.TargetSpec{
		Mode:  "permanent",
		Label: "target planeswalker that player controls or battle that player protects",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return (c.IsPlaneswalker() && c.Controller == player) ||
				(c.IsBattle() && c.ProtectorPlayerID == player)
		},
		Min: 1, Max: 1,
	}
}
