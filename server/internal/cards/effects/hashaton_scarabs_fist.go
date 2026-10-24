package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hashaton, Scarab's Fist — 1/3 Legendary Creature — Zombie Wizard
// for {W}{B}:
//
//	"Whenever you discard a creature card, you may pay {2}{U}. If you
//	 do, create a tapped token that's a copy of that card, except
//	 it's a 4/4 black Zombie."
//
// Three pieces, and the interesting one is the third:
//
//   - The discard trigger is EventDiscardCard, which carries the
//     discarded card, so the creature check reads the type line out
//     of the graveyard (the same route Mary Read takes).
//   - The optional cost is the MayPay primitive — "you may pay {N}.
//     If you do" is PayUnless with the consequence on the other
//     answer, and it now shares that prompt and that auto-tap.
//   - The token is a copy of a specific card rather than a template,
//     which is what CreateTokenCopy was built for. The copy carries
//     the discarded card's oracle ID, so its triggered abilities,
//     statics, keywords and activated abilities all resolve through
//     the ordinary catalog hooks — a copy of a catalog card really
//     behaves like the card.
//
// "Except it's a 4/4 black Zombie" REPLACES the creature types
// rather than adding to them (CR 707.9a): a discarded Human Wizard
// comes back as a Zombie, not a Zombie Human Wizard. Eternalize,
// which keeps the types, spells them out ("a 4/4 black Zombie Snake
// Wizard") — that contrast is the evidence. Supertypes survive, so a
// copy of a legendary creature is legendary.
//
// The copy is read from wherever the discarded card now sits. That is
// the printed result and not a simplification: a card exiled in
// response is still copied, because the copiable values are the printed
// ones and do not change by zone (CR 707.2), which is also what the
// last-known-information reading (CR 608.2h) gives. Since #762 a created
// token runs the ordinary entry pipeline, so a copied card's
// Spec.AsEnters clause fires on the token along with its ETB triggers.
//
// The discarded card's instance ID is a fact about what happened, so
// it rides item.Trigger.Event.CardID (#1223) rather than a value
// captured when the trigger fired.
func init() {
	Register(Spec{
		OracleID:     "db266661-f783-4907-9e52-6963eec05431",
		Name:         "Hashaton, Scarab's Fist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDiscardCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return discardedByYou(ev, source) && eventCardHasType(ev, g, "creature")
			},
			Key: "Hashaton, Scarab's Fist — pay {2}{U} to copy the discarded creature",
			// ADR 0126's amendment of 2026-10-06: a creature card
			// discarded may become a 4/4 token. The {2}{U} it asks for
			// is not in the declaration; the token is priced as any
			// token is, well under a 4/4, which leaves room for it.
			Purpose: game.Purpose{DiscardPayoff: &game.DiscardPayoff{Types: []string{"creature"}, Tokens: 1}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{2}{U}",
					Question: "Hashaton, Scarab's Fist — pay {2}{U} to create a tapped 4/4 black Zombie copy?",
					Then:     hashatonPaid,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// hashatonZombieException is the "except it's a tapped 4/4 black
// Zombie" clause. Tapped rides on the template because
// CreateTokenForEffect copies it wholesale — the token is never
// untapped on the battlefield, which is what "create a TAPPED token"
// says and what an OnETB tap would get subtly wrong.
func hashatonZombieException(t *game.Card) { hashatonZombieExceptionThrough(nil)(t) }

// hashatonZombieExceptionThrough is the exception with "black" and
// "Zombie" read through `reader`, the text of the ability that makes
// the copy (CR 612.2, #3984); nil reads the printed words.
func hashatonZombieExceptionThrough(reader *game.Card) func(t *game.Card) {
	return func(t *game.Card) {
		t.Power = 4
		t.Toughness = 4
		t.VariableToughness = false // a printed 4, not a `*` stand-in (#683)
		t.SetCopyExceptionColors(reader.TextColors("B")...)
		t.TypeLine = retypedTypeLine(t.TypeLine, reader.TextSubtypes("Zombie")...)
		t.Tapped = true
	}
}

// hashatonPaid is the paid half, keyed (#4264): the discarded card is the asking trigger's event.
var hashatonPaid = PayUnlessBranches("pay-unless/hashaton-scarabs-fist-zombie-copy",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return CreateTokenCopy{
			Controller: ctx.Controller(),
			Copy:       ctx.Trigger().Event.CardID,
			N:          1,
			Except:     hashatonZombieExceptionThrough(ctx.textReader()),
		}.Apply(ctx)
	}, nil)
