package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// embalm.go — #1221: embalm (CR 702.128) and eternalize (CR 702.129).
//
// Both are ACTIVATED ABILITIES that function only from their owner's
// GRAVEYARD and whose cost exiles the card that has them:
//
//	embalm      "{cost}, Exile this card from your graveyard: Create
//	            a token that's a copy of it, except it's white, it
//	            has no mana cost, and it's a Zombie in addition to
//	            its other types."                    CR 702.128a
//	eternalize  the same, except the token is black, it is 4/4, and
//	            it keeps its other types.            CR 702.129a
//
// One constructor pair over one body, because the two differ in the
// "except" clause and in nothing else — the same relationship Cycling
// and Typecycling have.
//
// The token is a copy of a card IN EXILE, which CreateTokenCopy has
// always been able to do (its doc names eternalize) — the source is
// read by instance ID through LookupCardForEffect, which finds a card
// in whatever zone holds it, and the activation cost has just put it
// in exile. That is also why the "except" clause edits a game.Card
// template rather than a PrintedValues: a token copy's copiable
// values ARE the template (ADR 0043 §1).
//
// CR 707.9b's ADD form, not CR 707.9a's set form, and the card says
// which: "a Zombie IN ADDITION TO its other types" keeps the Human
// Wizard. retypedTypeLine is the set form and is the wrong helper
// here — see addedSubtypeTypeLine below.

// Embalm is "Embalm <cost>" — CR 702.128a.
//
//	Activated: []ActivatedAbility{Embalm("{3}{W}")},
func Embalm(cost string) ActivatedAbility {
	return ActivatedAbility{
		Label: "Embalm " + cost + " (" + cost + ", Exile this card from your graveyard: " +
			"Create a token that's a copy of it, except it's white, it has no mana cost, " +
			"and it's a Zombie in addition to its other types. Activate only as a sorcery.)",
		Cost:         Plus(ManaCost(cost), ExileThis()),
		Zones:        []game.ZoneKind{game.ZoneGraveyard},
		SorcerySpeed: true,
		Effect:       embalmToken,
	}
}

// Eternalize is "Eternalize <cost>" — CR 702.129a. The token is a
// 4/4 BLACK Zombie and keeps everything else, which is why the two
// keywords are one body with two exception clauses.
//
//	Activated: []ActivatedAbility{Eternalize("{5}{B}{B}")},
func Eternalize(cost string) ActivatedAbility {
	ab := Embalm(cost)
	ab.Label = "Eternalize " + cost + " (" + cost + ", Exile this card from your graveyard: " +
		"Create a token that's a copy of it, except it's a 4/4 black Zombie " +
		"in addition to its other types and it has no mana cost. Activate only as a sorcery.)"
	ab.Effect = eternalizeToken
	return ab
}

// embalmToken and eternalizeToken are the two bodies, each one call
// into the shared one. Package-level and capture-free for the reason
// every ability body is.
func embalmToken(g *game.Game, item *game.StackItem) error {
	return zombieTokenCopyOfSource(g, item, embalmException)
}

func eternalizeToken(g *game.Game, item *game.StackItem) error {
	return zombieTokenCopyOfSource(g, item, eternalizeException)
}

// zombieTokenCopyOfSource creates one token copy of the ability's own
// source card, wherever it now sits — which is exile, because the
// activation cost put it there.
func zombieTokenCopyOfSource(g *game.Game, item *game.StackItem, except func(*game.Card)) error {
	return CreateTokenCopy{
		Controller: item.Controller,
		Copy:       item.SourceCardID,
		N:          1,
		Except:     except,
	}.Apply(NewContext(g, item))
}

// embalmException is "except it's white, it has no mana cost, and
// it's a Zombie in addition to its other types" (CR 702.128a).
//
// The P/T is untouched: an embalm token is the creature's own size.
func embalmException(t *game.Card) {
	t.Colors = []string{"W"}
	t.ManaCost = ""
	t.TypeLine = addedSubtypeTypeLine(t.TypeLine, "Zombie")
}

// eternalizeException is "except it's a 4/4 black Zombie in addition
// to its other types and it has no mana cost" (CR 702.129a).
func eternalizeException(t *game.Card) {
	t.Power = 4
	t.Toughness = 4
	// A printed 4, not a `*` stand-in (#683): an eternalized
	// Tarmogoyf is a 4/4 and its toughness is a real number.
	t.VariableToughness = false
	t.PrintedPTKnown = true
	t.Colors = []string{"B"}
	t.ManaCost = ""
	t.TypeLine = addedSubtypeTypeLine(t.TypeLine, "Zombie")
}

// addedSubtypeTypeLine rebuilds a type line with `subtype` ADDED to
// its subtypes rather than replacing them — CR 707.9b's "in addition
// to its other types", which is what embalm and eternalize print.
//
// retypedTypeLine (token_copy.go) is the other half of the same
// sentence, the SET form that CR 707.9a's bare "except it's a 4/4
// black Zombie" means, and the two are deliberately separate
// functions: picking the wrong one turns an embalmed Human Wizard
// into a plain Zombie, and no test of the copy machinery would
// notice.
//
// Already-present subtypes are not duplicated, and the subtype goes
// FIRST because that is how the printed reminder text reads ("a
// Zombie Human Wizard").
// `subtype` is "Zombie" at both call sites today — the parameter is
// there because the exception clauses read like the printed text when
// they name it, and because the function is CR 707.9b rather than a
// Zombie helper.
//
//nolint:unparam // see above
func addedSubtypeTypeLine(printed, subtype string) string {
	super, types, subs := game.ParseTypeLine(printed)
	for _, s := range subs {
		if strings.EqualFold(s, subtype) {
			return printed
		}
	}
	head := make([]string, 0, len(super)+len(types))
	head = append(head, super...)
	head = append(head, types...)
	line := strings.Join(head, " ")
	all := append([]string{subtype}, subs...)
	return line + " — " + strings.Join(all, " ")
}
