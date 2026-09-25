package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attacks_unblocked.go is CR 509.3's "whenever this creature attacks
// and isn't blocked" (#1279, ADR 0045 Decision 38).
//
// The trigger event is the defending player's COMPLETED block
// declaration, EventBlockersDeclared — one per defender per combat,
// emitted once their blocks are locked in. Before #1279 there was no
// such moment: an empty blocked record meant "not blocked" and "not
// asked yet" alike, so the trigger had nothing to fire on and every
// card printing it was left out.
//
// The rule's edges, and why each holds here:
//
//   - it triggers when NO creature is declared as a blocker for it —
//     the ability asks UnblockedAttackerForEffect, whose blocked record
//     is written by that same declaration before the event is emitted;
//   - it triggers for a creature that is attacking when blockers are
//     declared even if it was never declared as an attacker (it
//     entered attacking) — the question is "is it attacking this
//     defender", not "was it declared";
//   - it does NOT trigger for a creature put onto the battlefield
//     attacking after the declaration (ninjutsu's ninja), because that
//     creature was not there for the one event;
//   - it does NOT trigger for a blocked creature whose blockers have
//     all left combat — the record says blocked (CR 509.1h, #715).

// attacksAndIsNotBlocked is the AppliesTo half: `ev` is a completed
// block declaration by the player `source` is attacking, and nothing
// was declared blocking `source`.
//
// Caller holds the game lock (the harvester's contract).
func attacksAndIsNotBlocked(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventBlockersDeclared || source == nil {
		return false
	}
	if source.AttackingTarget == uuid.Nil {
		return false
	}
	if g.DefendingPlayerForAttackerForEffect(source.InstanceID) != ev.Actor {
		return false
	}
	return g.UnblockedAttackerForEffect(source.InstanceID)
}

// WhenAttacksAndIsNotBlocked builds the triggered ability. `effect`
// runs at resolution with the DEFENDING PLAYER — the player whose
// declaration fired it, captured as a value in Build — because every
// printed payoff of this trigger names that player ("defending player
// discards a card", "…gets a poison counter").
func WhenAttacksAndIsNotBlocked(label string, effect func(g *game.Game, item *game.StackItem, defender uuid.UUID) error) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventBlockersDeclared},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return attacksAndIsNotBlocked(ev, source, g)
		},
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			defender := ev.Actor
			return game.NewTriggeredItem(source, label, func(g *game.Game, item *game.StackItem) error {
				return effect(g, item, defender)
			})
		},
	}
}

// WhenAttacksAndIsNotBlockedEffect is WhenAttacksAndIsNotBlocked with
// its Effect declared on the row (ADR 0041 P9, #1497, tier 4-3): the
// defending player is read back from the item's carried trigger event
// (triggeringActor) rather than captured in a hand-written Build, so a
// table with the trigger waiting on the stack is a restore point.
func WhenAttacksAndIsNotBlockedEffect(label string, effect func(g *game.Game, item *game.StackItem, defender uuid.UUID) error) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventBlockersDeclared},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return attacksAndIsNotBlocked(ev, source, g)
		},
		Key: label,
		Effect: func(g *game.Game, item *game.StackItem) error {
			return effect(g, item, triggeringActor(item))
		},
	}
}

// defendingPlayerStillIn reports whether the defending player a trigger
// captured is still in the game, which every payoff here checks before
// touching them (CR 800.4a: a player who has left is not affected).
func defendingPlayerStillIn(g *game.Game, defender uuid.UUID) bool {
	p := g.PlayerByIDForEffect(defender)
	return p != nil && !p.Eliminated
}
