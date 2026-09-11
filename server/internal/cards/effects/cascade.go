package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cascade.go — S28: CR 702.85's keyword, as a constructor a card file
// can spell in one line.
//
//	Triggered: []game.TriggeredAbility{Cascade()},
//	Triggered: []game.TriggeredAbility{Cascade(), Cascade()},  // Maelstrom Wanderer
//
// The engine half is game/cascade.go — the exile loop, the random
// bottoming, the "you may cast it" prompt and the free-cast grant.
// This file is the two shapes the catalog needs: a spell that HAS
// cascade, and a permanent that GIVES cascade to other spells
// (Maelstrom Nexus, Imoti).

// Cascade is the keyword on the card that has it. Two copies in a
// card's Triggered list is "cascade, cascade", which is Maelstrom
// Wanderer exactly: two separate triggered abilities, and the player
// orders them on the stack.
//
// FromStack is what makes it work at all. Cascade triggers when the
// spell is CAST — while it is on the stack, several minutes of game
// time before it would reach the battlefield the harvester scans.
func Cascade() game.TriggeredAbility {
	return game.TriggeredAbility{
		FromStack: true,
		Watches:   []game.EventKind{game.EventCast},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID
		},
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			return cascadeItem(source.Name, ev.Actor, source.InstanceID, source.ManaValue())
		},
	}
}

// GrantsCascade builds the trigger for a permanent that gives cascade
// to spells its controller casts — Maelstrom Nexus ("the first spell
// you cast each turn"), Imoti ("spells you cast with mana value 6 or
// greater").
//
// `when` narrows it; it receives the SPELL as a value (already looked
// up out of the stack) alongside the granting permanent, so a card
// file never has to do the zone walk itself. Returning true for a
// spell the grantor's controller did not cast is not possible — the
// controller check is applied first, here, because every printed
// version of this clause says "you cast".
//
// Note the asymmetry with Cascade(): this trigger is on a BATTLEFIELD
// permanent, so it needs no FromStack. The cascading object is the
// spell; the triggering object is the enchantment watching it.
func GrantsCascade(label string, when func(spell game.Card, source *game.Card, g *game.Game) bool) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventCast},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if ev.Actor != source.Controller {
				return false
			}
			spell, ok := g.LookupCardForEffect(ev.CardID)
			if !ok {
				return false
			}
			return when == nil || when(spell, source, g)
		},
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			mv := 0
			if spell, ok := g.LookupCardForEffect(ev.CardID); ok {
				mv = spell.ManaValue()
			}
			return cascadeItem(label, ev.Actor, source.InstanceID, mv)
		},
	}
}

// cascadeItem is the stack item both shapes queue. `attribution` is
// the card ID the stack overlay and the event log hang the trigger
// off; `lessThan` is the cascading SPELL's mana value, captured at
// trigger time because by resolution the spell may have left the
// stack (countered, or resolved above the trigger — cascade goes on
// the stack above its own spell, so it always resolves first, but a
// Stifle-shaped answer is still a legal board state).
//
// The controller is ev.Actor, the player who cast the spell, not the
// source card's Controller field. For a spell cast out of its owner's
// own hand those agree; for one cast off an opponent's library under
// an impulse grant they do not, and cascade belongs to the caster.
func cascadeItem(name string, caster, attribution uuid.UUID, lessThan int) *game.StackItem {
	return &game.StackItem{
		Kind:         game.StackItemTriggered,
		Controller:   caster,
		Owner:        caster,
		SourceCardID: attribution,
		Label:        name + " — cascade",
		Effect: func(g *game.Game, item *game.StackItem) error {
			return g.CascadeForEffect(item.Controller, item.SourceCardID, lessThan)
		},
	}
}
