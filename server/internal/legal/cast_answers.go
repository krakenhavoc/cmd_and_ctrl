package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_answers.go — #2877, ADR 0142's amendment of 2026-10-30: the
// reader for what a SPELL answers. ADR 0142 §2 defined Purpose.Answers
// for a spell, a mode and an alternative cost and refused it there until
// the owner decided how a castable instant declared `value` (Opt,
// Brainstorm) should be classed. The owner chose option (a) on
// 2026-10-30: by default it does not stop smart autopass on an
// opponent's stack item, and a setting turns the stop back on. Combat
// and an opponent's end step are unchanged; they are the windows a
// value instant is cast in.
//
// The server says what the announcement answers; the client decides
// when that stops (ADR 0009). So the one output is Move.ValueOnly, set
// on a cast whose announcement is declared to answer nothing. Every
// other cast keeps today's meaning, an instant that stops you, and that
// includes a cast that declares nothing (ADR 0142 owner answer 3, ADR
// 0009 §3: a stop over a missed window).

// castAnswersOf is what one cast announcement declares it answers: the
// card's own declaration (`own`, game.CardPurposeOfPtr's Answers for
// the face being cast), a chosen mode's in place of the card's for that
// mode, or the alternative cost's when the cost declares one. Zero
// is "not declared", and any part of the announcement that declares
// nothing makes the whole of it undeclared:
//
//   - a mode that declares nothing is covered by the card's declaration,
//     and with none there it is undeclared;
//   - an alternative cost that declares answers names what the spell
//     does for that cost (overload turns Cyclonic Rift into a sweep). One
//     that declares nothing leaves the card's own effect, unless it says
//     the spell is something else for that cost: a purpose of its own
//     with no answers, "target" made "each" (overload), a clause list of
//     its own, a face-down cast or a bestowed Aura;
//   - a fused split spell's purpose is both halves' (game's plus), and an
//     undeclared half leaves the fused spell undeclared.
//
// A kicker adds nothing here. A card whose kicked effect answers
// something declares the union on the card (ADR 0142 sweep ruling 4),
// and effects.Register refuses Answers inside Purpose.Kicked.
func castAnswersOf(own game.Answers, modeSpec *game.ModeSpec, modes []int, offer *game.AlternativeCost) game.Answers {
	if offer != nil {
		if a := offer.Purpose.Answers; a != 0 {
			return a
		}
		if offerChangesTheSpell(offer) {
			return 0
		}
	}
	if modeSpec == nil || len(modes) == 0 {
		return own
	}
	var out game.Answers
	for _, i := range modes {
		a := own
		if i >= 0 && i < len(modeSpec.Options) {
			if m := modeSpec.Options[i].Purpose.Answers; m != 0 {
				a = m
			}
		}
		if a == 0 {
			return 0
		}
		out |= a
	}
	return out
}

// offerChangesTheSpell reports whether casting for this alternative cost
// makes the spell something other than the card's own effect, so the
// card's declaration does not speak for it.
func offerChangesTheSpell(offer *game.AlternativeCost) bool {
	return !offer.Purpose.IsZero() || offer.ClearsTargets || offer.Targets != nil ||
		offer.FaceDown != nil || offer.Bestow
}

// castValueOnly is Move.ValueOnly for one cast announcement: its
// declaration is exactly `value` (castAnswersOf), it targets nothing on
// the stack (a move that does is a counter to every reader), and its
// costs give up no creature. The last is read from the announcement,
// not the declaration, like ADR 0142 owner answer 4: sacrificing,
// exiling or returning a creature as a cost (Altar's Reap, Deadly
// Dispute, an evoke-style return) saves it from whatever is on the
// stack, so that cast is a sacrifice outlet whatever the card declares.
// effects.Register refuses `value` beside such a printed cost; this
// catches the rest (an alternative cost, a choice of what to give up).
func castValueOnly(g *game.Game, answers game.Answers, targetsStack bool, costIDs ...[]uuid.UUID) bool {
	if answers != game.AnswerValue || targetsStack {
		return false
	}
	for _, ids := range costIDs {
		for _, id := range ids {
			if c := battlefieldCardByID(g, id); c != nil && c.IsCreature() {
				return false
			}
		}
	}
	return true
}

// AdditionalCostAnswers is what a spell's additional cost alone answers,
// as CostAnswers is for an activated ability's: sacrificing or exiling a
// creature you control is a sacrifice outlet (Altar's Reap, Deadly
// Dispute; "sacrifice a land" or "a Food" is a price, not an answer). An
// either/or cost answers what any branch does. effects.Register reads
// it: a spell whose printed cost is an outlet may not declare `value`.
func AdditionalCostAnswers(c *game.AdditionalCost) game.Answers {
	if c == nil {
		return 0
	}
	var a game.Answers
	if sacrificeInteracts(c.Sacrifice) || exileCostInteracts(c.ExilePermanents) {
		a |= game.AnswerSacOutlet
	}
	for i := range c.Either {
		a |= AdditionalCostAnswers(&c.Either[i])
	}
	return a
}

// AlternativeCostAnswers is the same question for an alternative cost:
// sacrificing or returning a creature you control instead of paying
// (CR 118.9) is an outlet.
func AlternativeCostAnswers(c *game.AlternativeCost) game.Answers {
	if c == nil {
		return 0
	}
	if sacrificeInteracts(c.Sacrifice) || sacrificeInteracts(c.ReturnToHand) {
		return game.AnswerSacOutlet
	}
	return 0
}
