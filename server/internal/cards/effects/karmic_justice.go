package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Karmic Justice — Enchantment {2}{W} (Enchanting Tales):
//
//	"Whenever a spell or ability an opponent controls destroys a noncreature
//	 permanent you control, you may destroy target permanent that opponent
//	 controls."
//
// The trigger reads the exit's stamped cause (leave_cause.go): a battlefield
// exit by the word "destroy" (LeaveCauseDestroy) that a resolving spell or
// ability moved, whose controller (Event.LeaveController) is another player. State-
// based deaths, sacrifices, exile and a destruction a replacement turned into
// an exile do not count, and neither does a creature: it is judged as it last
// existed on the battlefield (CR 603.10a), so an animated land counts as a
// creature, and the permanent must have been under your control as it left.
// One trigger per permanent destroyed.
//
// "That opponent" is the controller of the spell or ability behind the
// destruction: the target clause is ControlledByThatEventLeaveController, so
// the legal set, the announcement and the resolution check agree (CR 608.2b).
// "You may" is the CR 603.5 optional trigger, asked as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9ea58595-fcb9-455c-9558-b6a3fb2b83c5",
		Name:         "Karmic Justice",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(Optional(game.TriggeredAbility{
				Watches:   []game.EventKind{game.EventZoneMove},
				AppliesTo: karmicJusticeSawADestruction,
				Key:       "Karmic Justice — you may destroy target permanent that opponent controls",
				Effect:    destroyFirstLegalTarget,
			}, "Karmic Justice — destroy target permanent that opponent controls?"),
				ControlledByThatEventLeaveController(TargetPermanent("target permanent that opponent controls"))),
		},
	})
}

// karmicJusticeSawADestruction is the trigger condition: a noncreature
// permanent the source's controller controlled was destroyed by a spell or
// ability another player controls.
func karmicJusticeSawADestruction(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.OldZone != game.ZoneBattlefield || ev.LeaveCause != game.LeaveCauseDestroy || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	if ev.LeaveController == uuid.Nil || ev.LeaveController == source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || leftAsType(ev, c, "creature") {
		return false
	}
	return leftUnderControlOf(ev, c) == source.Controller
}
