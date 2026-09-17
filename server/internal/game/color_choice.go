package game

import (
	"strings"

	"github.com/google/uuid"
)

// color_choice.go — "choose a color" (CR 105.4), issue #742.
//
// One prompt kind in two forms, because the printed phrase comes in
// two forms:
//
//   - STORED: "As this permanent enters, choose a color" — Coldsteel
//     Heart, Heraldic Banner, the Thriving lands and the Gates. The
//     answer is written onto the permanent's Card.ChosenColor and read
//     back for as long as it stays on the battlefield, by its mana
//     ability ("one mana of the chosen color") and its statics
//     ("creatures you control of the chosen color").
//   - AT RESOLUTION: "Choose a color. …" inside a spell or ability —
//     Wash Out, Oona, Selective Obliteration. Nothing is stored; the
//     answer is handed to a continuation that runs the rest of the
//     effect.
//
// The stored form copies S26's creature-type choice exactly, including
// its declared simplification: the prompt is queued from the
// permanent's AsEnters hook rather than by pausing the CR 614 entry
// pipeline, so the permanent is on the battlefield with an empty
// ChosenColor between entering and the answer arriving. Nothing can
// act in that window (an open PendingChoice stops priority), and the
// direction is the safe one: a mana ability reading an empty colour
// produces nothing, and an anthem reading one applies to nothing.
// creature_type_choice.go has the long form of that argument.
//
// CR 105.4: a player asked to choose a color picks exactly one of the
// five. Never "colorless" and never "multicolored" — which is why the
// option set is always a subset of AllColors, and why "a color other
// than blue" (Thriving Isle) is just a four-entry option list rather
// than a separate mechanism.

// PendingChoiceColor is the "choose a color" prompt. Answered with the
// same `{color: "G"}` payload a mana pick uses; the dispatcher routes
// the two by KIND, not by the presence of the field.
//
// ColorOptions carries the legal colours (a subset of AllColors).
const PendingChoiceColor PendingChoiceKind = "choose_color"

// EventColorChosen records a "choose a color" answer in the event log.
// `Label` carries the colour letter and `CardID` the card it was
// chosen for.
const EventColorChosen EventKind = "color_chosen"

// AllColors is the CR 105.1 colour set in WUBRG order, the order every
// colour picker renders.
var AllColors = []string{"W", "U", "B", "R", "G"}

// ColorsOtherThan is AllColors without `color` — the "choose a color
// other than blue" option set.
func ColorsOtherThan(color string) []string {
	out := make([]string, 0, len(AllColors))
	for _, c := range AllColors {
		if c != color {
			out = append(out, c)
		}
	}
	return out
}

// ColorName is the English name of a colour letter ("G" → "green"),
// for prompt labels and event text. Unknown input comes back as-is.
func ColorName(color string) string {
	switch color {
	case "W":
		return "white"
	case "U":
		return "blue"
	case "B":
		return "black"
	case "R":
		return "red"
	case "G":
		return "green"
	}
	return color
}

// normaliseColorOptions narrows `options` to real colours in WUBRG
// order with no duplicates; nil or empty means all five. A caller
// that passes "C" (colorless is not a color, CR 105.4) or junk gets
// it silently dropped rather than offered.
func normaliseColorOptions(options []string) []string {
	if len(options) == 0 {
		return append([]string(nil), AllColors...)
	}
	want := map[string]bool{}
	for _, o := range options {
		want[strings.ToUpper(o)] = true
	}
	out := make([]string, 0, len(AllColors))
	for _, c := range AllColors {
		if want[c] {
			out = append(out, c)
		}
	}
	return out
}

// chooseColorFrame is the continuation behind a resolution-time
// PendingChoiceColor. `then` receives the chosen colour and runs with
// g.mu held, so it may queue a further prompt (Selective
// Obliteration's next player). Counted in ContinuationCensus like
// every other frame.
type chooseColorFrame struct {
	then func(g *Game, color string) error
}

// QueueColorChoiceForEffect queues the STORED form: "as this permanent
// enters, choose a color". The answer lands on `source`'s ChosenColor.
// `options` nil means any of the five colours. Returns the choice ID.
//
// Caller must hold g.mu (an AsEnters hook does).
func (g *Game) QueueColorChoiceForEffect(chooser, source uuid.UUID, reason string, options []string) uuid.UUID {
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:         PendingChoiceColor,
		Chooser:      chooser,
		FromPlayer:   chooser,
		Count:        1,
		Source:       source,
		Reason:       reason,
		ColorOptions: normaliseColorOptions(options),
	})
}

// ColorPrompt is the queue-side description of a resolution-time
// "choose a color".
type ColorPrompt struct {
	// Chooser answers the prompt. Required.
	Chooser uuid.UUID
	// Source is the card asking. Empty is legal (test harnesses).
	Source uuid.UUID
	// Question is the prompt's header.
	Question string
	// Options are the legal colours; nil means all five.
	Options []string
	// Then receives the answer. Runs with g.mu held; may queue further
	// choices.
	Then func(g *Game, color string) error
}

// QueueColorChoiceThenForEffect queues the AT-RESOLUTION form: the
// answer is handed to p.Then and stored nowhere.
//
// Caller must hold g.mu.
func (g *Game) QueueColorChoiceThenForEffect(p ColorPrompt) uuid.UUID {
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:              PendingChoiceColor,
		Chooser:           p.Chooser,
		FromPlayer:        p.Chooser,
		Count:             1,
		Source:            p.Source,
		Reason:            p.Question,
		ColorOptions:      normaliseColorOptions(p.Options),
		chooseColorResume: &chooseColorFrame{then: p.Then},
	})
}

// ResolveColorChoice answers a PendingChoiceColor.
//
// The colour is validated against the choice's own option list (case
// insensitively) BEFORE dequeuing, so an illegal answer leaves the
// prompt open for another try.
//
// Stored form: the permanent is located live, like the creature-type
// resolver does — it may have left while the prompt was open, and a
// missing source is not an error (the choice is made and has nowhere
// to land). The layer version is bumped because the colour is an
// AppliesTo input to the source's statics (Heraldic Banner's anthem),
// and nothing else in this path emits an event the layer listener
// watches.
//
// Resolution form: the continuation runs; an error it returns is
// logged as EventEffectError and the prompt is gone either way, which
// is the #544 contract every other continuation resolver keeps.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveColorChoice(choiceID, chooserID uuid.UUID, color string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, choice := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	if choice.Kind != PendingChoiceColor {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	color = strings.ToUpper(strings.TrimSpace(color))
	legal := false
	for _, o := range choice.ColorOptions {
		if o == color {
			legal = true
			break
		}
	}
	if !legal {
		return ErrInvalidParam
	}
	frame := choice.chooseColorResume
	source := choice.Source
	g.dequeueChoiceLocked(idx)

	g.EmitEvent(Event{
		Kind:   EventColorChosen,
		Actor:  chooserID,
		CardID: source,
		Label:  color,
	})
	if frame != nil {
		if frame.then != nil {
			if err := frame.then(g, color); err != nil {
				g.EmitEvent(Event{
					Kind:     EventEffectError,
					Actor:    chooserID,
					Source:   source,
					ErrorMsg: err.Error(),
				})
			}
		}
	} else if i := findCardOnBattlefield(g, source); i >= 0 {
		g.Battlefield.Cards[i].ChosenColor = color
		g.layerVersion.Add(1)
	}
	g.runStateChecksLocked()
	return nil
}

// ChosenColorOf returns the colour chosen for the permanent `sourceID`
// currently on the battlefield, or "" when none has been chosen (or
// the permanent is gone).
//
// Caller must hold either lock.
func (g *Game) ChosenColorOf(sourceID uuid.UUID) string {
	if i := findCardOnBattlefield(g, sourceID); i >= 0 {
		return g.Battlefield.Cards[i].ChosenColor
	}
	return ""
}
