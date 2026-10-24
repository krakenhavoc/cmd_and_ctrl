package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nuka-Nuke Launcher — Artifact — Equipment {2}:
//
//	"Equipped creature gets +3/+0 and has intimidate.
//	 Whenever equipped creature attacks, until the end of defending
//	 player's next turn, that player gets two rad counters whenever
//	 they cast a spell.
//	 Equip {3}"
//
// #2809. The statics are the usual equipment pair (PumpAttached,
// GrantToAttached). The attack trigger schedules a CR 603.7 delayed
// triggered ability that lasts until the end of the DEFENDING player's
// next turn (UntilEndOfYourNextTurnDuration, which ends at that turn's
// cleanup step, CR 514.2) and fires on every spell that player casts
// until then (CR 603.7b, Repeats).
//
// The defending player is the player behind the attack's target
// (CR 508.5), fixed when the trigger resolves, and is data on the
// delayed trigger (Params.Player), so a table with it waiting is still
// a restore point. The ability is the Launcher's controller's
// (CR 603.7e), so that player gives the counters, and it outlives the
// Launcher and the creature (CR 603.7). Two attacks, or two Launchers,
// are two delayed triggers: a spell cast under both gets two sets of
// counters, as two such effects would.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "7fb88a5d-9b72-4da9-ade4-d09cadd7e1cb",
		Name:         "Nuka-Nuke Launcher",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(3, 0),
			GrantToAttached("intimidate"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverEnchantedCreatureAttacks("Nuka-Nuke Launcher — until the end of defending player's next turn, they get two rad counters whenever they cast a spell",
				nukaNukeLauncherSchedule),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}

// nukaNukeLauncherSchedule makes the delayed "whenever they cast a
// spell" trigger for the defending player.
func nukaNukeLauncherSchedule(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	defender := g.DefendingPlayerForAttackForEffect(item.Trigger.Event.Target)
	if p := g.PlayerByIDForEffect(defender); defender == uuid.Nil || p == nil || p.Eliminated {
		return nil
	}
	until := g.UntilEndOfYourNextTurnDuration(defender)
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   item.Controller,
		SourceCardID: item.SourceCardID,
		Label:        "Nuka-Nuke Launcher — that player gets two rad counters",
		On:           []game.EventKind{game.EventCast},
		Condition:    nukaNukeLauncherCastCondition,
		CondParams:   game.EffectParams{Player: defender},
		Body:         nukaNukeLauncherBody,
		Params:       game.EffectParams{Player: defender},
		Repeats:      true,
		Duration:     &until,
	})
	return nil
}

// nukaNukeLauncherCastCondition is "whenever they cast a spell": the
// cast's actor is the defending player the trigger was made for.
var nukaNukeLauncherCastCondition = game.DelayedCondition("nuka-nuke-launcher/defender-casts-a-spell",
	func(ev game.Event, _ *game.DelayedTrigger, _ *game.Game, p game.EffectParams) bool {
		return ev.Kind == game.EventCast && p.Player != uuid.Nil && ev.Actor == p.Player
	})

// nukaNukeLauncherBody gives the defending player two rad counters,
// given by the delayed trigger's controller.
var nukaNukeLauncherBody = game.DelayedBody("nuka-nuke-launcher/defender-gets-two-rad-counters",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		return playerGetsRadCounters(g, item.Controller, p.Player, 2)
	})
