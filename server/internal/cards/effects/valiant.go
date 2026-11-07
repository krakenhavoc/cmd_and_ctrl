package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// valiant.go — Bloomburrow's valiant (#5230):
//
//	"Valiant — Whenever this creature becomes the target of a spell or
//	 ability you control for the first time each turn, <effect>."
//
// Valiant is an ability word (CR 207.2c): it has no rules meaning of its
// own, so the whole ability is the trigger condition and the effect after
// it. A valiant card is written as
//
//	Triggered: []game.TriggeredAbility{
//	    Valiant("Heartfire Hero — Valiant: put a +1/+1 counter on it",
//	        Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
//	},
//
// and a targeted effect wraps the row in Targeting, as any other trigger
// does (Mouse Trapper's "tap target creature an opponent controls").
//
// # The trigger event
//
// game.EventBecomesTarget, the event Black Bolt, Bonecrusher Giant and the
// rest of the "becomes the target" family read. It is emitted once per
// target slot (CR 115.3) as the targets are announced (CR 601.2c for a
// spell, 603.3d for a triggered ability), and also when a spell or ability
// is copied (CR 707.10) or has its targets changed (CR 115.7), so the
// trigger goes on the stack above the thing that targeted and resolves
// first, as the Bloomburrow release notes say it does. Three filters:
//
//   - only this creature counts: the event's CardID is the source;
//   - "you control" is the spell or ability's controller, the event's
//     Actor, against the creature's controller now;
//   - "for the first time each turn" is below.
//
// # "For the first time each turn"
//
// The ruling is about the creature, not the ability: it triggers "if that
// creature hasn't yet been the target of a spell or ability you control
// this turn" (Bloomburrow release notes, 2024-07-26). So the record is the
// turn's event log, the same bounded g.EventsThisTurn() scan Zenith
// Chronicler's first multicolored spell reads, and NOT the per-object
// trigger tally (Game.TriggeredThisTurn) Hercules, Olympian Hero uses.
// The tally counts announcements, and two cases the rule covers never
// announce one:
//
//   - one spell that targets the creature in two slots (CR 115.3) emits
//     two events back to back; a targeted valiant ability's first trigger
//     is still waiting on its own target prompt when the second event is
//     harvested, so the tally would still read zero and fire it twice;
//   - a targeted valiant ability with no legal target is removed (CR
//     603.3d) and announces nothing, yet the creature has been targeted,
//     so a later spell that turn must not fire it.
//
// The scan stops at the event being asked about, so the answer is the
// same however often the harvester asks, and it reads only the turn
// boundary the log is cut at, so the next turn starts clean.
//
// # A new object
//
// A creature that leaves the battlefield and returns is a new object (CR
// 400.7), with no memory of having been targeted. The event log names a
// card by instance ID, which survives a zone change, so the scan forgets
// everything before the latest zone change of that card: a targeting
// event counts only when no EventZoneMove, EventETB or EventLTB for the
// same card came after it.

// Valiant is the valiant trigger: "Whenever this creature becomes the
// target of a spell or ability you control for the first time each turn,
// <effect>". The label is the stack label, "<card> — Valiant: <effect>".
func Valiant(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventBecomesTarget, ValiantTriggers, label, effect)
}

// ValiantTriggers is the valiant condition as a When, for a row that is
// built some other way than Valiant (an optional or modal one).
func ValiantTriggers(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || source == nil || ev.CardID == uuid.Nil || ev.CardID != source.InstanceID {
		return false
	}
	you := source.Controller
	if you == uuid.Nil || ev.Actor != you {
		return false
	}
	return !valiantTargetedEarlierThisTurn(g, ev, you)
}

// valiantTargetedEarlierThisTurn reports whether the object `ev` names had
// already become the target of a spell or ability `you` controlled this
// turn, before `ev`. A zone change of the card in between ends the object
// that was targeted (CR 400.7), so what came before it does not count.
func valiantTargetedEarlierThisTurn(g *game.Game, ev game.Event, you uuid.UUID) bool {
	targeted := false
	for _, prev := range g.EventsThisTurn() {
		if prev.Seq >= ev.Seq {
			break
		}
		if prev.CardID != ev.CardID {
			continue
		}
		switch prev.Kind {
		case game.EventZoneMove, game.EventETB, game.EventLTB:
			targeted = false
		case game.EventBecomesTarget:
			if prev.Actor == you {
				targeted = true
			}
		}
	}
	return targeted
}
