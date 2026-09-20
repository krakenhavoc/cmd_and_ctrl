package effects

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// color_choice.go — the card-side vocabulary for "choose a color"
// (CR 105.4, #742). The engine half is game/color_choice.go.
//
//	ChooseColorAsEnters(purpose, label)    "As this enters, choose a color."
//	ChooseColorOtherThanAsEnters(p, l, c)  "…choose a color other than blue."
//	ProducedChosenColor()                  "{T}: Add one mana of the chosen color."
//	ProducedColorOrChosen(c)               "{T}: Add {U} or one mana of the chosen color."
//	ChosenColorAnthem(p, t)                "Creatures you control of the chosen color get +p/+t."
//	ProducedOneColor(n)                    "Add N mana of any one color."
//	ChooseColorThen(…)                     "Choose a color. <rest of the effect>"
//
// Every reader of a stored colour treats "not chosen yet" as the
// weaker outcome — no mana, no anthem — never as "every colour".
//
// EVERY PROMPT DECLARES A PURPOSE (#780). CR 105.4 makes all five
// colours legal, so nothing about the prompt itself says which one the
// card wants — a chooser with no other information names its own main
// colour, which is right for Coldsteel Heart and bounces its own board
// for Wash Out. `game.ColorPurpose` is the card's one-word answer to
// "what happens to the colour I name", it is the first argument of
// every builder here, and `color_purpose_guard_test.go` fails the
// build for a prompt that does not pass one of the declared
// constants.

// ChooseColorAsEnters builds the `Spec.AsEnters` for "As this
// permanent enters, choose a color." `purpose` is what the permanent
// will do with the answer (#780) and `label` is the prompt header,
// normally the card's name. The answer lands on the permanent's
// Card.ChosenColor; see game/color_choice.go for why this is an ETB
// hook rather than a paused CR 614 replacement (S26's creature-type
// choice made the same call).
func ChooseColorAsEnters(purpose game.ColorPurpose, label string) func(*game.Card, *Context) error {
	return func(card *game.Card, ctx *Context) error {
		ctx.Game.QueueColorChoiceForEffect(card.Controller, card.InstanceID, label+" — choose a color", nil, purpose)
		return nil
	}
}

// ChooseColorOtherThanAsEnters is ChooseColorAsEnters for "choose a
// color other than <color>" — the Thriving lands and the Gates. The
// excluded colour is simply not offered.
func ChooseColorOtherThanAsEnters(purpose game.ColorPurpose, label, color string) func(*game.Card, *Context) error {
	options := game.ColorsOtherThan(color)
	return func(card *game.Card, ctx *Context) error {
		ctx.Game.QueueColorChoiceForEffect(card.Controller, card.InstanceID,
			label+" — choose a color other than "+game.ColorName(color), options, purpose)
		return nil
	}
}

// ChosenColorOf is the colour stored on the permanent `source`, or ""
// while none has been chosen. Read-only; safe under either lock.
func ChosenColorOf(g *game.Game, source uuid.UUID) string {
	return g.ChosenColorOf(source)
}

// ProducedChosenColor is the ProducedFunc for "{T}: Add one mana of
// the chosen color." Before a colour is chosen it produces nothing —
// the source still taps, and the player simply gets no mana, which is
// the weaker direction.
func ProducedChosenColor() func(*game.Game, uuid.UUID, uuid.UUID) string {
	return func(g *game.Game, _, source uuid.UUID) string {
		if c := g.ChosenColorOf(source); c != "" {
			return "{" + c + "}"
		}
		return ""
	}
}

// ProducedColorOrChosen is the ProducedFunc for "{T}: Add {U} or one
// mana of the chosen color." — the printed colour always, and the
// chosen one as a second option once it exists. A single-option result
// goes straight into the pool; a two-option one queues the ordinary
// mana pick.
func ProducedColorOrChosen(color string) func(*game.Game, uuid.UUID, uuid.UUID) string {
	color = strings.ToUpper(color)
	return func(g *game.Game, _, source uuid.UUID) string {
		chosen := g.ChosenColorOf(source)
		if chosen == "" || chosen == color {
			return "{" + color + "}"
		}
		return "{" + color + "|" + chosen + "}"
	}
}

// ChosenColorAnthem is "Creatures you control of the chosen color get
// +power/+toughness" (Heraldic Banner): layer 7c, applied to creatures
// the source's controller controls that have the colour stored on the
// source. An unchosen colour matches nothing.
func ChosenColorAnthem(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: CreatureYouControlOfChosenColor,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}

// CreatureYouControlOfChosenColor is the `StaticAbility.AppliesTo`
// body for "creatures you control of the chosen color".
func CreatureYouControlOfChosenColor(target *game.Card, _ *game.Game, source *game.Card) bool {
	if source.ChosenColor == "" || !target.IsCreature() {
		return false
	}
	return target.Controller == source.Controller && target.HasColor(source.ChosenColor)
}

// ProducedOneColor is the ProducedFunc for "Add N mana of any one
// color", where N is computed at activation (Mona Lisa's power, White
// Lotus Tile's largest tribe) — one pick, N tokens of the picked
// colour. N of zero or less produces nothing.
//
// It narrows to nothing: every printed card in this family says "any
// one color", with no commander-identity clause, so the ability that
// uses it leaves NarrowToCommanderIdentity off.
func ProducedOneColor(n func(g *game.Game, controller, source uuid.UUID) int) func(*game.Game, uuid.UUID, uuid.UUID) string {
	return func(g *game.Game, controller, source uuid.UUID) string {
		return OneColorOfAmount(n(g, controller, source))
	}
}

// OneColorOfAmount is the produced-mana string for "N mana of any one
// color": "{W3|U3|B3|R3|G3}" for three, "" for none. Also usable with
// AddManaForEffect from a spell or a trigger (Sanctum of Fruitful
// Harvest).
func OneColorOfAmount(n int) string {
	if n <= 0 {
		return ""
	}
	count := strconv.Itoa(n)
	parts := make([]string, 0, len(game.AllColors))
	for _, c := range game.AllColors {
		parts = append(parts, c+count)
	}
	return "{" + strings.Join(parts, "|") + "}"
}

// ChooseColorThen queues a resolution-time "Choose a color." for
// `chooser` and runs `then` with the answer. Nothing is stored. `then`
// runs with the game lock held, like every continuation — *ForEffect
// helpers only.
//
// `purpose` is #780's declaration: what the continuation is going to do
// to the colour it is handed. It leads the argument list because it is
// a property of the QUESTION, and a reviewer should see it before the
// wording.
func ChooseColorThen(purpose game.ColorPurpose, g *game.Game, chooser, source uuid.UUID, question string, then func(g *game.Game, color string) error) {
	g.QueueColorChoiceThenForEffect(game.ColorPrompt{
		Chooser:  chooser,
		Source:   source,
		Question: question,
		Purpose:  purpose,
		Then:     then,
	})
}

// AnyCombinationOfColors is the produced-mana string for "N mana in
// any combination of colors" (Selvala, Heart of the Wilds; Chromatic
// Orrery): N INDEPENDENT any-colour slots, so the controller answers
// N picks and may answer them differently. Zero or fewer produces
// nothing.
//
// The counterpart to OneColorOfAmount, and the difference between
// them is the whole reason the produced-mana grammar carries a
// per-colour amount. "Add three mana in any combination of colors"
// is "{W|U|B|R|G}{W|U|B|R|G}{W|U|B|R|G}" and can pay {W}{U}{B};
// "add three mana of any one color" is "{W3|U3|B3|R3|G3}", one pick
// minting three of the same. Reaching for the wrong one ships a card
// that looks right and fixes or splits the colours wrongly.
func AnyCombinationOfColors(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("{"+strings.Join(game.AllColors, "|")+"}", n)
}

// ProducedAnyCombinationOfColors is the ProducedFunc for "Add X mana
// in any combination of colors", where X is computed at activation
// (Selvala's greatest power among creatures you control).
//
// No commander-identity narrowing, for OneColorOfAmount's reason:
// every printed card in this family says "any combination of colors"
// with no identity clause, so the ability leaves
// NarrowToCommanderIdentity off.
func ProducedAnyCombinationOfColors(n func(g *game.Game, controller, source uuid.UUID) int) func(*game.Game, uuid.UUID, uuid.UUID) string {
	return func(g *game.Game, controller, source uuid.UUID) string {
		return AnyCombinationOfColors(n(g, controller, source))
	}
}
