package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jin-Gitaxias, Progress Tyrant — Legendary Creature — Phyrexian
// Praetor {5}{U}{U}, 5/5 (slice 296-m):
//
//	"Whenever you cast an artifact, instant, or sorcery spell, copy
//	 that spell. You may choose new targets for the copy. This ability
//	 triggers only once each turn.
//	 Whenever an opponent casts an artifact, instant, or sorcery
//	 spell, counter that spell. This ability triggers only once each
//	 turn. (A copy of a permanent spell becomes a token.)"
//
// Two EventCast triggers, gated the same way: the caster
// (ev.Actor, you or an opponent), the spell's own printed type
// (looked up off the stack by ev.CardID, since the cast event fires
// with the spell already there), and Morbid Opportunist's
// once-each-turn tally (b11TriggeredThisTurn, keyed on On's own
// label) so a second artifact/instant/sorcery spell in a turn does
// nothing on either half.
//
// The copy is CopySpell with ChooseNewTargets, CR 707.10's own
// "you may choose new targets for the copy" prompt — a copy of a
// permanent spell becomes a token per CR 706.9, which CopySpell
// already handles for every other copy effect in the catalog. The
// counter is CounterTarget on the cast spell directly: this ability
// does not target ("counter THAT spell", not "counter target
// spell"), so there is no legality re-check beyond the spell still
// being on the stack.
//
// No simplification.
const jinCopyLabel = "Jin-Gitaxias, Progress Tyrant — copy that spell"
const jinCounterLabel = "Jin-Gitaxias, Progress Tyrant — counter that spell"

func init() {
	Register(Spec{
		OracleID:     "f5daadc1-98ff-480a-82bb-fe7bfaa7b60e",
		Name:         "Jin-Gitaxias, Progress Tyrant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller || !jinMatchesSpellTypes(g, ev.CardID) {
					return false
				}
				return !b11TriggeredThisTurn(g, source.InstanceID, jinCopyLabel)
			}, jinCopyLabel, jinCopyThatSpell),
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor == source.Controller || !jinMatchesSpellTypes(g, ev.CardID) {
					return false
				}
				return !b11TriggeredThisTurn(g, source.InstanceID, jinCounterLabel)
			}, jinCounterLabel, jinCounterThatSpell),
		},
	})
}

func jinCopyThatSpell(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return CopySpell{
		StackID:          ctx.Trigger().Event.CardID,
		ChooseNewTargets: true,
	}.Apply(ctx)
}

func jinCounterThatSpell(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return CounterTarget{StackID: ctx.Trigger().Event.CardID}.Apply(ctx)
}

// jinMatchesSpellTypes is "an artifact, instant, or sorcery spell",
// read off the stack by the cast event's card id.
func jinMatchesSpellTypes(g *game.Game, cardID uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(cardID)
	if !ok {
		return false
	}
	return c.IsArtifact() || c.IsInstant() || c.IsSorcery()
}
