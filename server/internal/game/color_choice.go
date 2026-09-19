package game

import (
	"log/slog"
	"strings"
	"sync"

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

// ColorPurpose is what the card will DO with the colour it is asking
// for. It is declared by the card at the point it asks, it is public
// information (anybody can read the card), and it exists because CR
// 105.4 makes every one of the five colours a LEGAL answer — so the
// only thing that separates a good answer from a terrible one is what
// the effect does next.
//
// #780: without it every automated chooser answered "my main colour",
// which is right for Coldsteel Heart and is a self-inflicted board
// wipe for Wash Out. The engine does not read this field; it carries
// it to the wire, where a policy (or a client hint) can.
//
// Keep the set SMALL. A purpose is a shape of question, not a card: the
// bar for a new one is that no existing purpose gives a sane answer for
// a whole family of cards.
type ColorPurpose string

const (
	// ColorForMana — "add one mana of the chosen color" (Coldsteel
	// Heart, the Thriving lands, the Gates). The colour is a mana
	// source, so the right answer is whatever the chooser needs to
	// cast things with.
	ColorForMana ColorPurpose = "mana"

	// ColorForBenefit — the chosen colour is the one that gets HELPED,
	// or, on Selective Obliteration, the one that SURVIVES. Heraldic
	// Banner's anthem and "exile each permanent unless it's only the
	// color its controller chose" are the same question from the
	// chooser's side: name the colour you want to keep.
	ColorForBenefit ColorPurpose = "benefit"

	// ColorForHarm — everything of the chosen colour is punished, the
	// chooser's own permanents included (Wash Out). The right answer
	// maximises what the opposition loses net of what the chooser
	// does.
	ColorForHarm ColorPurpose = "harm"

	// ColorForFilter — the colour selects which of a set of unknown
	// or opposing cards the effect acts on (Oona, Queen of the Fae).
	// Nothing of the chooser's is at stake either way.
	ColorForFilter ColorPurpose = "filter"

	// ColorForProtection — the chosen colour is the one being
	// defended AGAINST (Mother of Runes, Story Circle, the Circles of
	// Protection). The right answer is the colour of whatever is
	// about to hurt you.
	ColorForProtection ColorPurpose = "protect"
)

// AllColorPurposes is every declared purpose, for the catalog guard
// and for anything that has to validate one off the wire.
var AllColorPurposes = []ColorPurpose{
	ColorForMana, ColorForBenefit, ColorForHarm, ColorForFilter, ColorForProtection,
}

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
//
// Dropping EVERY entry is the case that is not silent. A prompt with no
// options is not a narrow colour prompt, it is an unanswerable one: the
// enumerator offers the seat no answers and the choice blocks the
// table, which is the #499 / #618 wedge, and the client's
// colorPromptAnswerable refuses to open a picker for it — so a
// miswritten card ("C" alone, a typo'd letter) stalls the game instead
// of asking a bad question. CR 105.4 says the answer is one of the
// five, so the five are the honest fallback, exactly as they are for
// the empty input above. The warning is the breadcrumb that says a card
// file is wrong, in #844's posture: play continues, once per process.
// Found by #986's ordering work.
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
	if len(out) == 0 {
		emptyColorOptionsOnce.Do(func() {
			slog.Warn("a choose_color prompt narrowed to no colours: offering all five",
				"options", options, "rule", "CR 105.4", "issue", 986)
		})
		return append([]string(nil), AllColors...)
	}
	return out
}

// emptyColorOptionsOnce keeps the warning above to one line per
// process. It names a CARD BUG, so it is the same line every time and
// repeating it per prompt would bury the rest of the log.
var emptyColorOptionsOnce sync.Once

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
func (g *Game) QueueColorChoiceForEffect(chooser, source uuid.UUID, reason string, options []string, purpose ColorPurpose) uuid.UUID {
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:         PendingChoiceColor,
		Chooser:      chooser,
		FromPlayer:   chooser,
		Count:        1,
		Source:       source,
		Reason:       reason,
		ColorOptions: normaliseColorOptions(options),
		ColorPurpose: purpose,
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
	// Purpose is what the card does with the answer (#780). Required
	// in practice: the catalog guard fails a prompt that leaves it
	// empty, and an empty one falls back to the mana-fixing rule.
	Purpose ColorPurpose
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
		ColorPurpose:      p.Purpose,
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
