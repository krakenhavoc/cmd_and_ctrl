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

// KeywordCascade and KeywordStorm are the machine-readable names the
// two keyword constructors stamp on game.TriggeredAbility.Keyword
// (#1258), which is what cards/coverage reads to ask whether a card
// has the keyword. Neither is a canonicalKeywords token: both are
// catalog constructors, the pattern ADR 0014's 2026-09-24 amendment
// keeps for a keyword trigger that is never granted by a static.
const (
	KeywordCascade = "cascade"
	KeywordStorm   = "storm"
)

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
		Keyword:   KeywordCascade,
		FromStack: true,
		Key:       cascadeKey,
		Watches:   []game.EventKind{game.EventCast},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID
		},
		// A fill-in Build (ADR 0041 P9): the label, the caster and the
		// spell's mana value are facts of the moment the spell was
		// cast; the effect is the row's.
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			mv, _ := g.ManaValueForEffect(*source)
			return cascadeItem(source.Name, ev.Actor, source.InstanceID, mv)
		},
		Effect: cascadeEffect,
	}
}

// cascadeKey is the Key of every cascade row — the name a restored
// item checks its row by (ADR 0041 P9). Each item's own label names
// the card ("Bloodbraid Elf — cascade").
const cascadeKey = "cascade"

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
		// The granted cascade is still cascade (#1258): a caveat on
		// Maelstrom Nexus that said otherwise would be false.
		Keyword: KeywordCascade,
		Key:     label,
		Effect:  cascadeEffect,
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
				mv, _ = g.ManaValueForEffect(spell)
			}
			return cascadeItem(label, ev.Actor, source.InstanceID, mv)
		},
	}
}

// cascadeItem is the stack item both shapes queue, built by their
// fill-in Builds. `attribution` is the card ID the stack overlay and
// the event log hang the trigger off; `lessThan` is the cascading
// SPELL's mana value, captured at trigger time because by resolution
// the spell may have left the stack (countered, or resolved above the
// trigger — cascade goes on the stack above its own spell, so it
// always resolves first, but a Stifle-shaped answer is still a legal
// board state). Both shapes read it with
// game.(*Game).ManaValueForEffect, so an X spell's limit counts the X
// chosen for it (CR 202.3e). It rides the item as Params.Amount, and
// the row's Effect (cascadeEffect) reads it back — data, so a cascade
// waiting on the stack is a restore point (ADR 0041 P9).
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
		Params:       game.EffectParams{Amount: lessThan},
	}
}

// cascadeEffect is cascade's resolution, read entirely off the item.
func cascadeEffect(g *game.Game, item *game.StackItem) error {
	return g.CascadeForEffect(item.Controller, item.SourceCardID, item.Params.Amount)
}
