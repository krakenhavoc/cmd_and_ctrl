package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// storm.go — S46: CR 702.40's keyword, as a constructor a card file
// can spell in one line.
//
//	Triggered: []game.TriggeredAbility{Storm()},
//	Triggered: []game.TriggeredAbility{Storm(), Storm()},  // CR 702.40b
//
// The engine half is game/storm.go (the count) and game/turn_tally.go
// (TurnTally.Casts, the turn's cast order it is read from). The copies
// are the CR 707.10 primitive in spell_copy.go, unchanged. See
// ADR 0086.
//
// STORM IS CR 702.40 IN THE PINNED EDITION, not 702.39. The rules this
// tree cites are the ones effective August 7, 2026 (AGENTS.md §6), in
// which 702.39 is PROVOKE. Issue #1238 and the roadmap comments that
// fed it carry the older number; every citation here is checked
// against the pinned text.

// Storm is the keyword on the card that has it:
//
//	"When you cast this spell, copy it for each other spell that was
//	 cast before it this turn. If the spell has any targets, you may
//	 choose new targets for any of the copies." (CR 702.40a)
//
// Two copies in a card's Triggered list is two separate triggered
// abilities and two separate counts, which is CR 702.40b exactly. No
// card in paper prints storm twice; the constructor supports it
// because the rule says it, and because Strionic Resonator copying a
// storm trigger reaches the same place.
//
// FromStack is what makes it work at all, for cascade's reason: storm
// triggers when the spell is CAST — while it is on the stack, and
// never anywhere the battlefield scan can see it.
//
// Three things this deliberately does NOT do, each of them the printed
// text rather than a simplification:
//
//   - It does not filter by type, colour or controller. CR 702.40a
//     says "each other spell", and that is every player's — the
//     opponent's Swords to Plowshares counts. (Thousand-Year Storm's
//     "each other instant and sorcery spell YOU'VE cast" is the
//     narrowed version and keeps its own helper.)
//   - It does not care what happened to those spells afterwards. A
//     spell that was cast and then countered was still cast, which is
//     why the count is read off the recorded cast ORDER rather than
//     by looking each card up.
//   - It does not need the card to print "you may choose new targets".
//     ChooseNewTargets is set unconditionally: the engine skips the
//     CR 707.10c offer for a copy whose original chose no target
//     (itemHasChosenTarget), so Empty the Warrens — which prints no
//     such clause because it has no targets — is one line like the
//     rest and gets no prompts.
func Storm() game.TriggeredAbility {
	return game.TriggeredAbility{
		Keyword:   KeywordStorm,
		FromStack: true,
		Key:       stormKey,
		Watches:   []game.EventKind{game.EventCast},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID
		},
		// A fill-in Build (ADR 0041 P9): the label and the caster are
		// facts of the cast; the effect is the row's.
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			return stormItem(source.Name, ev.Actor, source.InstanceID)
		},
		Effect: stormEffect,
	}
}

// stormKey is the Key of every storm row — the name a restored item
// checks its row by (ADR 0041 P9).
const stormKey = "storm"

// stormItem is the stack item the keyword queues.
//
// It carries exactly two values, and both are immutable: the spell's
// instance ID (the item's SourceCardID) and the name the label is built
// from. Nothing is captured in a closure — the row's Effect,
// stormEffect, reads the spell off the item, so a storm trigger waiting
// on the stack is a restore point (ADR 0041 P9). The COUNT is not
// carried — it is read as the item resolves, off the same cast order
// the whole turn is recorded in (ADR 0086 Decision 2, CR 608.2h). A
// spell cast in response to this trigger is appended after the storm
// spell and cannot move its index, so the number is the same whenever
// it is asked; reading it at resolution is what makes that a property
// rather than a coincidence.
//
// The controller is ev.Actor, the player who CAST the spell, not the
// card's Controller field — cascade's reason: for a spell cast out of
// its owner's own hand the two agree, and for one cast off an
// opponent's library under an impulse grant they do not. The copies
// belong to the caster (CR 707.10b).
//
// A count of zero is an ordinary resolution, not an early return: the
// engine has already announced "storm count 0" by then, which is the
// line that tells a reader the trigger fired and found nothing to
// copy.
func stormItem(name string, caster, spell uuid.UUID) *game.StackItem {
	return &game.StackItem{
		Kind:         game.StackItemTriggered,
		Controller:   caster,
		Owner:        caster,
		SourceCardID: spell,
		Label:        name + " — storm",
	}
}

// stormEffect is storm's resolution: copy the spell the item names
// once for each spell cast before it this turn.
func stormEffect(g *game.Game, item *game.StackItem) error {
	spell := item.SourceCardID
	n := g.StormCountForEffect(item.Controller, spell)
	if n <= 0 {
		return nil
	}
	return CopySpell{
		StackID:          spell,
		Controller:       item.Controller,
		Count:            n,
		ChooseNewTargets: true,
		// CR 608.2h, #1255: the trigger names the spell and does not
		// target it, so a storm spell countered in response to its own
		// trigger (CR 113.7a keeps the trigger) is still copied, from
		// last-known information.
		FromLastKnown: true,
	}.Apply(NewContext(g, item))
}
