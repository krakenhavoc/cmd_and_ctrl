package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choose_option.go — the card-side vocabulary for "As this enters,
// choose <A> or <B>" where A and B are ANCHOR WORDS printed on the card
// (CR 614.12, #1572). The engine half is game/choose_option.go.
//
//	AsEnters:  ChooseOptionAsEnters("Frostcliff Siege", "Jeskai", "Temur"),
//	Triggered: []game.TriggeredAbility{
//	    WhenChosen("Jeskai", WheneverOneOrMore…(…)),       // • Jeskai — …
//	},
//	Static: []game.StaticAbility{
//	    StaticWhenChosen("Temur", TribalAnthem(…, 1, 0)),   // • Temur — …
//	},
//
// The ability printed after an anchor word carries the ADR 0071 gate
// ChosenIs(word), so it exists only while that word is the permanent's
// chosen option. A card file never writes "if the chosen option is
// Temur" inside an AppliesTo or an Apply: the Temur line does not exist
// on a Jeskai Siege, which is the stronger statement and the one the
// Siege rulings make (each Siege has ONE of its two listed abilities).
//
// Before the controller answers, NEITHER line exists. That is the only
// safe reading of an empty answer, and it is the gate's, not a card's.

// ChooseOptionAsEnters builds the `Spec.AsEnters` for "As this
// permanent enters, choose <options[0]> or <options[1]>." `label` is
// the prompt header, normally the card's name; `options` are the
// anchor words in printed order — the order the prompt offers them in,
// and the first is the one a bot with no opinion takes.
//
// The answer lands on the permanent's Card.ChosenOption; see
// game/choose_option.go for why this is an AsEnters hook rather than a
// paused CR 614 replacement (S26's creature-type choice made the same
// call, and #742, #980 and #1210 after it).
func ChooseOptionAsEnters(label string, options ...string) func(*game.Card, *Context) error {
	words := append([]string(nil), options...)
	question := label + " — choose " + joinOr(words)
	return func(card *game.Card, ctx *Context) error {
		ctx.Game.QueueChooseOptionAsEntersForEffect(card.Controller, card.InstanceID, question, words)
		return nil
	}
}

// ChosenIs is the CR 614.12 anchor-word gate: the ability exists while
// `option` is the permanent's chosen option. Pair it with the same
// spelling passed to ChooseOptionAsEnters —
// TestEveryAnchorWordGateIsOffered fails the build for a gate whose
// word the card's own prompt never offers, which would otherwise ship
// that line switched off forever.
func ChosenIs(option string) game.Designation { return game.ChosenOptionIs(option) }

// WhenChosen stamps an anchor-word gate onto a triggered ability built
// with any of the ordinary trigger constructors, so a bullet reads as
// one thing:
//
//	WhenChosen("Dragons", AtYourUpkeep("Palace Siege — drain 2", …))
func WhenChosen(option string, t game.TriggeredAbility) game.TriggeredAbility {
	t.ActiveWhen = ChosenIs(option)
	return t
}

// StaticWhenChosen is WhenChosen for a static ability — the Temur
// line of Frostcliff Siege, "Creatures you control get +1/+0 and have
// trample and haste."
func StaticWhenChosen(option string, s game.StaticAbility) game.StaticAbility {
	s.ActiveWhen = ChosenIs(option)
	return s
}

// DoublerWhenChosen is WhenChosen for a game.TriggerDoubler — Windcrag
// Siege's Mardu line, "If a creature attacking causes a triggered
// ability of a permanent you control to trigger, that ability
// triggers an additional time" (#1647). A TriggerDoubler is not an
// ability-list entry (TriggersForCard / StaticAbilitiesForCard never
// see it), so it carries its own ActiveWhen rather than reaching the
// ordinary gate through one of those accessors.
func DoublerWhenChosen(option string, d game.TriggerDoubler) game.TriggerDoubler {
	d.ActiveWhen = ChosenIs(option)
	return d
}

// ChosenOptionOf is the option stored on the permanent `source`, or ""
// while none has been chosen. Read-only; safe under either lock. The
// gate is the ordinary reader; this is for an effect that has to say
// the word (none in the catalog yet does).
func ChosenOptionOf(g *game.Game, source uuid.UUID) string {
	return g.ChosenOptionOf(source)
}

// joinOr renders an option list the way the card prints it: "Khans or
// Dragons", "A, B or C".
func joinOr(words []string) string {
	switch len(words) {
	case 0:
		return ""
	case 1:
		return words[0]
	}
	return strings.Join(words[:len(words)-1], ", ") + " or " + words[len(words)-1]
}
