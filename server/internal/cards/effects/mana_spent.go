package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_spent.go — the catalog-side vocabulary for #761's record of
// what mana paid for a spell: converge (CR 702.86), sunburst (CR
// 702.44), adamant (CR 207.2c) and "if no mana was spent to cast it".
//
// Each of these is one read of game.PaidCost through effects.Context,
// and each is written once here rather than per card. A card file says
// what the card says — ctx.ColorsSpentCount(), AdamantSpent("R", 3) —
// and never reaches for the record itself.
//
// Sunburst is the one that is NOT here. CR 702.44a is an "enters with
// counters" clause, so since #1002 it is declared as one —
// SunburstCounters in entry_counters.go — and the engine reads the
// paid-cost record off the resolving stack item while the CR 614 entry
// window is open, rather than the card reading it a beat earlier in
// OnResolve.
//
// The one rule they all inherit, from ADR 0068 §3: a payment the
// engine WAIVED (permissive mode, a strict-mode override) answers
// "unknown", and every reader here takes unknown as the weaker
// answer. Converge counts no colours, adamant does not turn on, and
// "if no mana was spent" is false. None of these helpers has to say
// so; PaidCost already does.

// AdamantSpent is adamant's condition (CR 207.2c): "if at least three
// <colour> mana was spent to cast this spell". An ability word, so the
// counting lives on each card — this is the shared reading of it.
//
// False for a payment the engine waived, which is the
// weaker-than-printed direction: the spell does the printed thing and
// not the bonus.
func AdamantSpent(ctx *Context, color string, n int) bool {
	return ctx.ManaSpentOfColor(color) >= n
}

// NoManaWasSpentToCast reports "if no mana was spent to cast it" for a
// spell already on the stack — the read a CAST TRIGGER makes, where
// the trigger's source is some other permanent (Vexing Bauble) and the
// spell is named by the event.
//
// For a spell the stack item's ID IS the card's instance ID, so
// EventCast.CardID is the key. True only when the engine KNOWS nothing
// was spent: a cascade or "without paying its mana cost" free cast, a
// {0} alternative cost, or a copy of a spell (CR 707.10 — mana is not
// an object, so nothing was spent to cast the copy).
//
// A permissive or forced cast answers FALSE. The player paid something
// the engine did not see, and a punisher that fired on it would
// counter half the spells cast at a table running the human default.
//
// Read-only, and it runs under g.mu: a trigger's AppliesTo already
// does.
func NoManaWasSpentToCast(g *game.Game, spellID uuid.UUID) bool {
	return g.StackItemPaidForEffect(spellID).NoManaSpent()
}

// ProducedPerCounterRemoved is the produced-mana string for a mana
// ability that adds one slot per counter its cost removed —
// Mage-Ring Network's "Add {C} for each storage counter removed this
// way" is ProducedPerCounterRemoved("{C}") (#789).
//
// `slot` is one brace token; it is repeated once per counter. Removing
// nothing produces nothing, which is the printed behaviour of a
// variable removal that removed no counters — the land still taps.
func ProducedPerCounterRemoved(slot string) func(*game.Game, uuid.UUID, uuid.UUID, game.PaidCost) string {
	return func(_ *game.Game, _, _ uuid.UUID, paid game.PaidCost) string {
		out := ""
		for i := 0; i < paid.CountersRemoved; i++ {
			out += slot
		}
		return out
	}
}

// ManaAbilityNotUsedThisTurn is "Activate only once each turn" for a
// MANA ability (Ramos, Dragon Engine).
//
// It counts this permanent's mana-ability activations in the turn's
// own slice of the event log rather than in a new per-turn map,
// because the fact is already recorded: ActivateManaAbility emits one
// EventManaAbilityActivated per activation, with Source set to the
// permanent. A new map would be new turn-scoped state to clone,
// snapshot, classify and flush, for a question the log already
// answers.
//
// LIMIT, stated because it is real: the count is per PERMANENT, not
// per ability index — the event carries no index. On a card with one
// mana ability, which is Ramos and every other card that prints this
// clause today, the two are the same thing. A card with two mana
// abilities, only one of them once-per-turn, would need the index on
// the event first.
//
// Read-only under g.mu, like every activation condition.
func ManaAbilityNotUsedThisTurn() ActivationCondition {
	return func(g *game.Game, _ uuid.UUID, source uuid.UUID) bool {
		for _, ev := range g.EventsThisTurn() {
			if ev.Kind == game.EventManaAbilityActivated && ev.Source == source {
				return false
			}
		}
		return true
	}
}
