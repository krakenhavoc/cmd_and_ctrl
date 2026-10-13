package game

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// text_box.go — exchanging text boxes (CR 612.5, CR 701.12h; ADR 0160).
//
// # What a text box is here
//
// A card's rules text is not text in this engine. It is the catalog
// entry its KEY resolves to — every trigger, activated and mana
// ability, static, replacement, cost modifier and printed keyword a
// Spec declares — plus the printed keywords the deck importer stamped
// on Card.Keywords. So a text box is exactly those two things,
// TextBoxValues, and exchanging two text boxes swaps them and nothing
// else: name, types, P/T, mana cost and colour are the flat printed
// fields, which nothing here touches (CR 612.5).
//
// # Where it is applied
//
// Not in the layer pass. The pass gathers each permanent's statics off
// its key before layer 1 runs, so a text box that changes WHICH statics
// exist has to be decided before the pass: the reason ADR 0043 gives
// for materialising duration copies first, and done the same way.
// materialiseTextBoxesLocked writes the winning value onto Card.TextBox
// at the top of every recompute (right after the duration copies, so
// layer 3 follows layer 1), when a record is registered, and when the
// sweep drops one. catalogKeyOf answers TextBox.Key first, and every
// catalog reader already goes through catalogKeyOf.
//
// # The record
//
// One ScopedEffect per creature, carrying one ModExchangeText mod with
// the text box that creature now has, BY VALUE (CR 701.12h: "the rules
// text of each permanent becomes the PREVIOUS rules text of the
// other"). Each is pinned to its creature, so a creature that leaves
// takes its record with it (CR 400.7) and the other keeps its new text
// for the rest of the effect's duration.

// ModExchangeText is "<affected> has this text box" (layer 3, CR 612.5):
// one half of an exchange of text boxes. Reads TextBox, which holds
// exactly ONE TextBoxValues. Not a layer-pass operation — see the file
// comment — so its reader is readerText.
const ModExchangeText ModKind = "textBox"

// TextBoxValues is one object's rules text as this engine models it
// (ADR 0160 Decision 1). Plain data: a snapshot carries it and an undo
// shares it.
//
// IMMUTABLE once built. Card.TextBox points at one, and Clone shares
// the pointer the way it shares Card.printed.
type TextBoxValues struct {
	// Key is the catalog key the abilities resolve through — the other
	// object's catalogKeyOf when the exchange happened, face suffix and
	// CR 707.9a copy grants included. Empty is a real value: a face-down
	// creature's rules text (CR 708.2a), or an uncatalogued card's, is
	// "nothing the engine automates".
	Key string `json:"key,omitempty"`
	// Keywords are the importer's printed keywords that were part of
	// that text (Card.Keywords).
	Keywords []string `json:"keywords,omitempty"`
	// From names the object the text was printed on, for the wire.
	From string `json:"from,omitempty"`
}

func (v TextBoxValues) clone() TextBoxValues {
	v.Keywords = copyStrings(v.Keywords)
	return v
}

func (v TextBoxValues) equal(o TextBoxValues) bool {
	return v.Key == o.Key && v.From == o.From && slices.Equal(v.Keywords, o.Keywords)
}

// ExchangeTextMod is one half of an exchange: "<affected> has the text
// box `v`". `v` is cloned.
func ExchangeTextMod(v TextBoxValues) Mod {
	return Mod{Kind: ModExchangeText, TextBox: []TextBoxValues{v.clone()}}
}

// textBoxModProblem is the registration and restore check: exactly one
// text box on a textBox mod, and none on any other kind. "" when sound.
func textBoxModProblem(m Mod) string {
	if m.Kind == ModExchangeText {
		if len(m.TextBox) != 1 {
			return fmt.Sprintf("a textBox mod carries exactly one text box, got %d", len(m.TextBox))
		}
		return ""
	}
	if len(m.TextBox) != 0 {
		return fmt.Sprintf("mod %q carries a text box, which only textBox reads", m.Kind)
	}
	return ""
}

func cloneTextBoxSlice(in []TextBoxValues) []TextBoxValues {
	if len(in) == 0 {
		return nil
	}
	out := make([]TextBoxValues, len(in))
	for i := range in {
		out[i] = in[i].clone()
	}
	return out
}

// textKeywords is the importer's printed keywords of the rules text this
// object has right now: the exchanged text's under a record, its own
// otherwise.
func (c *Card) textKeywords() []string {
	if c.TextBox != nil {
		return c.TextBox.Keywords
	}
	return c.Keywords
}

// TextBoxFrom names the object whose text box this permanent has, or ""
// when it has its own (ADR 0160 Decision 7).
func (c *Card) TextBoxFrom() string {
	if c.TextBox == nil {
		return ""
	}
	if c.TextBox.From == "" {
		return "another creature"
	}
	return c.TextBox.From
}

// currentTextBoxOf is the text box `c` has right now: the exchanged one
// under a record, else its own — its catalog key with no record in the
// way, and its own import-road keywords.
func currentTextBoxOf(c *Card) TextBoxValues {
	if c.TextBox != nil {
		return c.TextBox.clone()
	}
	return TextBoxValues{
		Key:      printedTextKeyOf(c),
		Keywords: copyStrings(c.Keywords),
		From:     c.Name,
	}
}

// textBoxFor returns the text box the records say `c` has, and false
// when none applies. Every record carries a whole text box, so applying
// them in timestamp order (CR 613.7) leaves exactly the latest; ties
// keep registration order, as the layer sort does.
func textBoxFor(c *Card, records []*ScopedEffect) (*TextBoxValues, bool) {
	var (
		best  *TextBoxValues
		bestT int64
	)
	for _, e := range records {
		if !affectsObject(e.Affected, c) {
			continue
		}
		for i := range e.Mods {
			m := &e.Mods[i]
			if m.Kind != ModExchangeText || len(m.TextBox) != 1 {
				continue
			}
			if best == nil || e.Timestamp >= bestT {
				best, bestT = &m.TextBox[0], e.Timestamp
			}
		}
	}
	return best, best != nil
}

// affectsObject is affectedPredicate without the closure.
func affectsObject(set []AffectedObject, target *Card) bool {
	for _, a := range set {
		if !memberMatches(a, target) {
			continue
		}
		if a.Controller != uuid.Nil && target.Controller != a.Controller {
			continue
		}
		return true
	}
	return false
}

// materialiseTextBoxesLocked brings every battlefield permanent's
// Card.TextBox in line with the textBox records — layer 3 of CR 613.1c
// for this effect class. Idempotent: a permanent already carrying the
// right value is not touched, so a recompute that changes nothing
// writes nothing. Bumps the layer version when it writes, because the
// key every ability is read through just changed.
//
// Caller must hold g.mu (write).
func (g *Game) materialiseTextBoxesLocked() {
	if g.Battlefield == nil {
		return
	}
	var records []*ScopedEffect
	for i := range g.ScopedEffects {
		if scopedEffectHasMod(&g.ScopedEffects[i], ModExchangeText) {
			records = append(records, &g.ScopedEffects[i])
		}
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if len(records) == 0 && c.TextBox == nil {
			continue
		}
		want, ok := textBoxFor(c, records)
		if !ok {
			if c.TextBox != nil {
				c.TextBox = nil
				g.layerVersion.Add(1)
			}
			continue
		}
		if c.TextBox != nil && c.TextBox.equal(*want) {
			continue
		}
		v := want.clone()
		c.TextBox = &v
		g.layerVersion.Add(1)
	}
}

// ExchangeTextBoxesForEffect exchanges the text boxes of the permanents
// `a` and `b` for duration `d` (CR 612.5, CR 701.12h). Each gets the
// text box the other has right now — its CURRENT one, so a creature
// already under an exchange passes on what it was given. Each half is
// a record pinned to its creature: a creature that leaves takes its
// half with it, and the other keeps its new text until `d` ends.
//
// CR 701.12a: if the whole exchange can't happen, none of it does —
// both must be on the battlefield, and they must be two different
// objects. Reports whether the exchange happened.
//
// Caller must hold g.mu (write); effects call it from the resolution
// frame, which does.
func (g *Game) ExchangeTextBoxesForEffect(sourceID, a, b uuid.UUID, d Duration, label string) bool {
	if a == b || a == uuid.Nil || b == uuid.Nil {
		return false
	}
	ca, okA := g.battlefieldCardLocked(a)
	cb, okB := g.battlefieldCardLocked(b)
	if !okA || !okB {
		return false
	}
	// The records already registered decide what each has now.
	g.materialiseTextBoxesLocked()
	textA, textB := currentTextBoxOf(ca), currentTextBoxOf(cb)
	ts := timeNowUnixNano()
	pinA := g.PinnedObjectsLocked(a)
	pinB := g.PinnedObjectsLocked(b)
	if len(pinA) != 1 || len(pinB) != 1 {
		return false
	}
	if !g.appendScopedEffectLocked(sourceID, pinA, ScopeNone, uuid.Nil, []Mod{ExchangeTextMod(textB)}, g.PinnedTo(d, a), label, ts) {
		return false
	}
	if !g.appendScopedEffectLocked(sourceID, pinB, ScopeNone, uuid.Nil, []Mod{ExchangeTextMod(textA)}, g.PinnedTo(d, b), label, ts) {
		// Unreachable — b was checked above — but half an exchange is
		// exactly what CR 701.12a forbids, so take the first half back.
		g.ScopedEffects = g.ScopedEffects[: len(g.ScopedEffects)-1 : len(g.ScopedEffects)-1]
		return false
	}
	g.materialiseTextBoxesLocked()
	return true
}

// settleEntryTextExchangeLocked registers an "as this enters, you may
// exchange its text box and another creature's" exchange (Deadpool,
// Trading Card; ADR 0160 Decision 6) once the zone move has stamped the
// entering permanent's object identity, and before EventETB, so the
// permanent's enters triggers are the gained text's (CR 603.6a).
//
// Caller must hold g.mu (write).
func (g *Game) settleEntryTextExchangeLocked(ev *ReplacementEvent, entered uuid.UUID) {
	if ev == nil || ev.exchangeTextWith == uuid.Nil {
		return
	}
	other := ev.exchangeTextWith
	ev.exchangeTextWith = uuid.Nil
	c, ok := g.battlefieldCardLocked(entered)
	if !ok {
		return
	}
	label := fmt.Sprintf("%s — exchanged text boxes", c.Name)
	g.ExchangeTextBoxesForEffect(entered, entered, other, IndefiniteDuration(), label)
}
