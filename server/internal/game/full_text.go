package game

import (
	"reflect"
)

// full_text.go — "this creature has the full text of that card"
// (CR 612.6, Volrath's Shapeshifter; #6233, ADR 0160's amendment of
// 2026-11-25).
//
// # What full text changes
//
// CR 612.6: "One card (Volrath's Shapeshifter) states that an object
// has the 'full text' of another object. This changes not just the
// text that appears in the object's text box and type line, but also
// changes the text that represents its name, mana cost, color
// indicator, power, and toughness." So for as long as the condition
// holds the permanent IS, in every printed respect, the other card:
// its oracle identity (so every catalog hook keys on the other card's
// abilities), name, mana cost, colours, type line, keywords and P/T.
// The one thing it keeps is what the printed text says it keeps:
// Volrath's Shapeshifter "has the text '{2}: Discard a card.'", a
// grant the card names (FullTextStatic.Keep).
//
// # Where it is applied
//
// The values come from the same place a copy's do, and land in the
// same place: the flat printed fields, MATERIALISED before the layer
// pass, for ADR 0043 Decision 1's reason (the oracle ID and the
// printed P/T are not on a Characteristic). Unlike a duration copy
// there is no record to register: the static is read continuously
// from the permanent's catalog entry and the top card of its
// controller's graveyard, on every layer recompute. Every graveyard
// arrival and departure already bumps the layer version
// (layer_listener.go, #1117), so the overlay follows the graveyard
// with no event of its own.
//
// It is a text-changing effect (CR 612), not a copy effect, so it
// sits ON TOP of the copies (CR 613.1a then 613.1c): the materialiser
// lifts it off first, runs the copies, and puts it back last. It is
// not copiable (CR 707.2): a Clone copying a Shapeshifter that has the
// text of a Grizzly Bears copies Volrath's Shapeshifter, which is what
// CopiableValuesOf reads through FullTextBase.
//
// # The baseline
//
// `Card.FullTextBase` is the permanent's printed values UNDER the
// overlay — its own, or its copies' — and it is what the lift puts
// back. It is not PrintedSelf: a permanent with full text is not a
// copy (IsCopy stays false), and CR 400.7's battlefield-leave revert
// puts FullTextBase back too (restorePrintedSelf).

// FullTextStatic is a card's printed "As long as the top card of your
// graveyard is a creature card, this creature has the full text of that
// card and has the text …" (Volrath's Shapeshifter, CR 612.6). Pure
// data, read through the catalog.
type FullTextStatic struct {
	// Keep is the grant (effects.Spec.Grants key) the permanent keeps
	// while it has the other card's full text: Volrath's Shapeshifter's
	// "{2}: Discard a card." Empty keeps nothing.
	Keep string
}

// fullTextStaticOf returns the full-text static the catalog entry under
// `key` prints, or nil.
func fullTextStaticOf(key string) *FullTextStatic {
	d := catalogDef(key)
	if d == nil {
		return nil
	}
	return d.FullTextOfTopGraveyardCreature
}

// liftFullTextLocked takes a full-text overlay off the permanent,
// putting FullTextBase back on its flat printed fields, and returns the
// values it had WITH the overlay so the caller can tell whether
// anything changed. nil, having done nothing, when there was none.
//
// It does not bump the layer version: materialiseFullTextLocked does,
// once it knows whether the values the permanent ends up with differ.
func (c *Card) liftFullTextLocked() *PrintedValues {
	if c.FullTextBase == nil {
		return nil
	}
	prev := printedValuesOf(*c)
	c.setPrintedValues(*c.FullTextBase)
	c.FullTextBase = nil
	return &prev
}

// fullTextSourceLocked returns the values `c` has the full text of
// right now: the top card of its controller's graveyard when the
// permanent's catalog entry prints the static and that card is a
// creature card (CR 404.2's "single face-up pile" has a top). Read
// after the copies, so a Clone of Volrath's Shapeshifter has the
// static and reads its own controller's graveyard.
//
// Caller must hold g.mu.
func (g *Game) fullTextSourceLocked(c *Card) (PrintedValues, *FullTextStatic, bool) {
	if c.FaceDownIsPermanent() {
		return PrintedValues{}, nil, false
	}
	st := fullTextStaticOf(CatalogKey(*c))
	if st == nil {
		return PrintedValues{}, nil, false
	}
	p := g.PlayerByIDForEffect(c.Controller)
	if p == nil || p.Graveyard == nil || len(p.Graveyard.Cards) == 0 {
		return PrintedValues{}, nil, false
	}
	top := p.Graveyard.Cards[len(p.Graveyard.Cards)-1]
	if !top.IsCreature() {
		return PrintedValues{}, nil, false
	}
	// A card in a graveyard is never a copy, so its copiable values are
	// its own printed ones; the face that is up, single-faced (a
	// permanent with a DFC's text cannot transform, CR 712.9's reason).
	return SingleFacedCopiableValuesOf(top), st, true
}

// materialiseFullTextLocked puts the full-text overlay on `c` when its
// static applies, after the copies have been materialised. `prev` is
// what liftFullTextLocked returned: the values the permanent had before
// this pass. The layer version is bumped, and the card-carried ability
// slices dropped, only when the permanent's printed values changed.
//
// Caller must hold g.mu (write).
func (g *Game) materialiseFullTextLocked(c *Card, prev *PrintedValues) {
	src, st, ok := g.fullTextSourceLocked(c)
	if !ok && prev == nil {
		return
	}
	if ok {
		base := printedValuesOf(*c)
		want := src.Clone()
		want.GrantAbility(st.Keep)
		want = want.withTokenStatus(IsTokenPtr(c))
		c.FullTextBase = &base
		c.setPrintedValues(want)
	}
	if prev != nil && reflect.DeepEqual(*prev, printedValuesOf(*c)) {
		return
	}
	c.clearCarriedAbilitySlices()
	g.layerVersion.Add(1)
}
