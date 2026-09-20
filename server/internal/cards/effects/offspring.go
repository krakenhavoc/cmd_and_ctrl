package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// offspring.go — the offspring keyword, both halves.
//
// Its own file rather than a tail on additional_cost.go or
// triggers_common.go, per the convention enters_tapped.go set:
// concurrent card batches collide on shared files, and offspring is
// one mechanic whose cost half and trigger half are meaningless
// apart.
//
// "Offspring {cost}" reads:
//
//	"You may pay an additional {cost} as you cast this spell. If you
//	 do, when this creature enters, create a 1/1 token copy of it."
//
// So it is KICKER with a fixed payoff, and it is deliberately built
// out of the same two pieces Gatekeeper of Malakir is: an optional
// additional cost (ADR 0073), and an entry trigger whose CR 603.4
// intervening if reads the record the resolution path carried onto
// the permanent. What offspring adds is that the payoff is printed
// into the keyword rather than into the card, which is exactly why it
// belongs here and not in a card file — the next offspring creature
// is two lines.
//
// It is NOT an alternative cost. The offspring cost is paid ON TOP of
// the mana cost, not instead of it, so a card that reached for
// game.AlternativeCost would cast for {B} rather than {2}{B}{B}.

// OffspringKey is the optional cost's stable identity, the string the
// entry trigger asks for. It is not one of the three keys the ENGINE
// reads (kicker, multikicker, buyback) — nothing in the engine has to
// know what offspring is, because the whole payoff is the trigger
// below.
const OffspringKey = "offspring"

// Offspring is the cost half — "Offspring {cost}", declared in
// Spec.OptionalCosts:
//
//	OptionalCosts: []game.AdditionalCost{Offspring("{B}")},
//
// Use it rather than a hand-rolled game.AdditionalCost, for the
// reason Kicker and Buyback have constructors: the Key is what the
// trigger reads back, and a literal with the wrong key (or none)
// compiles into a creature that never makes its token.
func Offspring(mana string) game.AdditionalCost {
	return game.AdditionalCost{
		Optional: true,
		Key:      OffspringKey,
		ManaCost: mana,
		Label:    "Offspring " + mana,
	}
}

// OffspringToken is the trigger half — the printed "when this
// creature enters, create a 1/1 token copy of it", declared beside
// the card's own triggers:
//
//	Triggered: []game.TriggeredAbility{OffspringToken("Darkstar Augur"), …},
//
// Two things it gets from being written once:
//
//   - The "if you do" is a CR 603.4 INTERVENING IF, checked when the
//     ability would trigger rather than when it resolves. A creature
//     cast without the offspring cost puts nothing on the stack, so
//     nobody is given a window to respond to a trigger that is not
//     happening.
//   - The token is a real copy (CR 707.2), so it carries the copied
//     card's oracle ID and therefore its triggers, statics, keywords
//     and abilities — a token copy of a flier flies. What it does not
//     carry is the ORIGINAL's cast record: a token has no Provenance,
//     so the intervening if is false for it and the token cannot
//     make a token of its own. That is both the rule and the reason
//     the copy is safe to take from the catalog rather than from a
//     hand-written template.
//
// The token is copied from the permanent itself, wherever it is by
// the time the trigger resolves. A creature killed in response is
// still copied, off its printed values in the graveyard, which is
// last-known information doing what it should.
func OffspringToken(cardName string) game.TriggeredAbility {
	return On(game.EventETB, AllOf(Self, offspringWasPaid),
		cardName+" — create a 1/1 token copy of it",
		func(g *game.Game, item *game.StackItem) error {
			return CreateTokenCopy{
				Controller: item.Controller,
				Copy:       item.SourceCardID,
				N:          1,
				Except:     offspringOneOne,
			}.Apply(NewContext(g, item))
		})
}

// offspringWasPaid is the intervening if: was this permanent's spell
// cast with the offspring cost paid? Read off the permanent's own
// record (ADR 0073 §5), so it is false for a token, false for a
// reanimated copy, and false for one put onto the battlefield without
// being cast at all.
func offspringWasPaid(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return game.OptionalCostTimesPaid(*source, source.Provenance.OptionalCosts, OffspringKey) > 0
}

// offspringOneOne is the keyword's "except it's 1/1" clause. It sets
// the P/T outright rather than modifying it, which is what makes an
// offspring token of a 6/6 a 1/1 — and it clears the variable-P/T bit
// with it, because a fixed 1 is not a stand-in for a characteristic-
// defining toughness however the original printed it (#683).
func offspringOneOne(t *game.Card) {
	t.Power = 1
	t.Toughness = 1
	t.VariableToughness = false
}
