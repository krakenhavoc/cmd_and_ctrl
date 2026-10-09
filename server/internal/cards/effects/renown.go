package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// renown.go — the card-facing half of CR 702.112, renown and the
// renowned designation (#2049, ADR 0071 amendment 2026-10-09). The
// engine half is game/renown.go: the "renown N" keyword token, its
// trigger, Card.Renowned and the DesignationRenowned gate.
//
// Renown itself needs nothing here: declare it and the engine does the
// rest, so a creature whose only text is keywords and renown (Topan
// Freeblade, Citadel Castellan, …) needs no card file at all.
//
//	PrintedKeywords: []string{"renown 1"},
//
// What a card file adds is what reads the designation:
//
//	Static:    []game.StaticAbility{RenownedKeywords("menace")},         // "As long as this creature is renowned, it has menace"
//	Triggered: []game.TriggeredAbility{WhenACreatureYouControlBecomesRenowned("…", effect)},
//	if g.IsRenowned(id) { … }                                            // "If it's renowned, untap it" on a target
//	if ThisWasRenowned(ctx) { … }                                        // "if this creature is renowned", re-checked on resolution

// Renowned is the CR 702.112b gate: this ability exists while the
// permanent is renowned — "As long as this creature is renowned, …".
// Once renowned a permanent stays renowned until it leaves the
// battlefield.
func Renowned() game.Designation { return game.RenownedGate() }

// RenownedKeywords is "As long as this creature is renowned, it has
// [keywords]" — a layer-6 self-grant behind the Renowned gate, the
// renown twin of MonstrousKeywords.
func RenownedKeywords(keywords ...string) game.StaticAbility {
	kws := append([]string(nil), keywords...)
	return game.StaticAbility{
		Layer:      game.Layer6Ability,
		ActiveWhen: Renowned(),
		AppliesTo:  selfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			appendKeywordsTo(c, kws)
		},
	}
}

// WhenACreatureYouControlBecomesRenowned is "Whenever a creature you
// control becomes renowned, …" (Valeron Wardens). It watches the one
// event a renown trigger emits, which fires only on the resolution
// that actually made the creature renowned. "You control" is read as
// the creature became renowned: the event's Actor is its controller
// then.
func WhenACreatureYouControlBecomesRenowned(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventBecameRenowned, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if source == nil || ev.Actor != source.Controller {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.IsCreature()
	}, label, effect)
}

// ThisIsRenowned is the trigger half of an intervening "if this
// creature is renowned" (CR 603.4): the source permanent is renowned
// as the event happens. Pair it with ThisWasRenowned in the effect.
func ThisIsRenowned(source *game.Card) bool {
	return source != nil && source.Renowned
}

// ThisWasRenowned is the resolution half of the same intervening if:
// the ability's source is renowned as the ability resolves, read as it
// last existed on the battlefield if it has left by then (CR 608.2h).
// A source that left and came back is a new object, which is read as
// the old one last existed, never as the new one.
func ThisWasRenowned(ctx *Context) bool {
	info, ok := ctx.SourcePermanent()
	return ok && info.Renowned
}
