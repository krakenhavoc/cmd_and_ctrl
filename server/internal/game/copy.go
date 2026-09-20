package game

// copy.go — CR 707 copy effects (Clone, Phyrexian Metamorph, Spark
// Double, Sakashima the Impostor). The S16.5 half of #159 that a
// real card finally asked for, and the fix for #335.
//
// # What a copy effect actually changes
//
// CR 613.1a puts copy effects in LAYER 1, before everything else,
// and CR 707.2 says what they copy: the COPIABLE VALUES — the
// printed characteristics, as already modified by other copy
// effects, plus the values set by "as this enters" and face-down
// effects. NOT counters, NOT other layers' continuous effects, NOT
// status (tapped, attacking, damage). A Clone of a creature wearing
// a Glorious Anthem is the printed 2/2, not the 3/3.
//
// # Where it lives in this engine
//
// The layer engine (layers.go) recomputes a `Characteristic` from
// the printed baseline on every relevant event. It would have been
// possible to express the copy as a Layer-1 `ContinuousEffect` that
// overwrites that Characteristic — and it would have been WRONG in
// a way that is invisible until you play the card, because a
// Characteristic carries Name / types / colours / P/T / abilities
// and nothing else. A copy has to bring:
//
//   - the oracle ID, or `CatalogKey` keeps resolving the Clone's own
//     (empty) catalog entry and the copy of Blood Artist has no
//     trigger, the copy of Lord of Atlantis grants nothing, and the
//     copy of Sol Ring taps for nothing;
//   - the printed P/T as PRINTED data, or the 0/0 Clone dies to the
//     CR 704.5f state-based action in the window between entering
//     and the next layer recompute (the recompute deliberately does
//     not run inside the SBA loop — ADR 0012 §1);
//   - the mana cost, colour identity, keywords, starting loyalty,
//     layout and faces, which no Characteristic field carries.
//
// So a copy effect here MATERIALISES one card's printed values onto
// another card, exactly the way ADR 0034's `SetFace` materialises
// one face's printed values onto a multi-face card. That is the
// same operation with a different source, and it is deliberately
// written to look like it. The copy becomes the layer engine's
// printed baseline, which is what CR 613.1a asks for: layer 1 is
// applied, and layers 2-7 then run on its result.
//
// Layer1Copy stays in `layerOrder` for the effect class this does
// NOT cover: a copy effect with a DURATION and a timestamp of its
// own (Mirage Mirror's "becomes a copy until end of turn",
// Cytoshape). Those have to be re-applied on every recompute and
// ordered against other layer-1 effects; an entry copy never does,
// because it is settled once, as the permanent enters, and never
// changes again while the permanent is on the battlefield.
//
// # Why the original values are kept
//
// CR 400.7: a permanent that changes zones becomes a new object.
// The copy effect applied to the PERMANENT, not to the card, so a
// Clone that dies is a card named Clone in its owner's graveyard —
// not a second Llanowar Elves. `Card.PrintedSelf` holds what the
// card reverts to, and the layer listener restores it on the
// battlefield-leave path (layer_listener.go), the same place it
// clears the effective-characteristic cache.

// PrintedValues is a portable snapshot of one card's printed
// characteristics — everything a copy effect brings across
// (CR 707.2). Pure data: no closures, no pointers, so it
// serialises into the game snapshot and deep-copies for undo.
//
// It deliberately does NOT carry `ManaAbilities` /
// `ActivatedAbilities`. Those hold closures, which cannot be
// serialised, and they do not need to be: for an ordinary card they
// are looked up from the catalog by oracle ID, which this snapshot
// does carry, so they follow the copy for free. `applyCopy` copies
// the card-carried slices separately for the one case where they
// are the only source of truth — copying a TOKEN, which has no
// oracle ID for the catalog to key on.
type PrintedValues struct {
	// OracleID is the copied card's catalog identity. This is the
	// single most load-bearing field in the struct: every catalog
	// hook (ETB triggers, static abilities, replacements, activated
	// and mana abilities, printed keywords) keys on it through
	// CatalogKey, so copying it is what makes the copy behave like
	// the thing it copied rather than looking like it.
	OracleID string

	// TokenKey is the OTHER catalog identity — a token template's
	// synthetic key (#521, game/token_key.go). Copied for exactly the
	// reason OracleID is: it is what the ability hooks key on, so a
	// copy of a Treasure token is a Treasure that still sacrifices
	// for mana.
	//
	// Empty for every printed card, which is what makes CR 707.2's
	// ordinary direction fall out: a token copying a card takes that
	// card's oracle ID and this blank, and CatalogKey prefers the
	// oracle ID. The two are never both set on one object.
	TokenKey string

	// ScryfallID is the printing identity, which is what the client
	// resolves card art from. Copied so a Clone of Llanowar Elves
	// shows Llanowar Elves.
	ScryfallID string

	Name     string
	TypeLine string
	ManaCost string

	Colors        []string
	ColorIdentity []string
	ProducedMana  []string
	Keywords      []string

	// GrantedAbilities are the catalog keys of ability bundles a copy
	// effect's "except" clause GRANTED to this object (CR 707.9a).
	// Names, not closures: the abilities themselves are static
	// catalog data, registered by the card that grants them and
	// looked up through the composite catalog key (copy_grants.go).
	//
	// It sits in the copiable values rather than on the replacement
	// that made the copy because CR 707.9a's second sentence makes a
	// granted ability copiable in its own right: a Clone copying a
	// Phantasmal Image gets the Image's sacrifice trigger, and it
	// gets it by copying this slice like any other printed value.
	GrantedAbilities []string

	Power           int
	Toughness       int
	StartingLoyalty int

	// VariableToughness travels with Toughness: a Clone of a `*`
	// creature copies the 0 stand-in, so it copies the bit that says
	// the 0 is a stand-in too.
	VariableToughness bool

	// PrintedPTKnown travels for the same reason VariableToughness
	// does, one sign the other way: a token copy of a living weapon
	// Germ copies a real printed 0/0, so it copies the bit that says
	// the 0 is real and dies to CR 704.5f like the original.
	PrintedPTKnown bool

	// Layout / Faces / ActiveFace are ADR 0034's multi-face data.
	// Copied wholesale: copying the front face of a transform card
	// gives a permanent that can still transform, which is right,
	// and copying a modal DFC's land back gives a land.
	Layout     string
	Faces      []Face
	ActiveFace int
}

// CopiableValuesOf returns the values a copy effect would take from
// `src` right now (CR 707.2).
//
// It reads the card's CURRENT printed fields, not `PrintedSelf`,
// and that is the rule rather than an accident: copiable values are
// "the printed values, as modified by other copy effects", so a
// Clone copying a Clone copies what that Clone became, not the word
// "Clone". Because `applyCopy` writes the copied values into the
// flat printed fields, reading them back here is that rule.
//
// Nothing from any other layer is read: no counters, no anthem, no
// type-change, no status. That is why this reads `src.Power` and
// not `src.Effective().Power`.
func CopiableValuesOf(src Card) PrintedValues {
	return PrintedValues{
		OracleID:          src.OracleID,
		TokenKey:          src.TokenKey,
		ScryfallID:        src.ScryfallID,
		Name:              src.Name,
		TypeLine:          src.TypeLine,
		ManaCost:          src.ManaCost,
		Colors:            copyStringSlice(src.Colors),
		ColorIdentity:     copyStringSlice(src.ColorIdentity),
		ProducedMana:      copyStringSlice(src.ProducedMana),
		Keywords:          copyStringSlice(src.Keywords),
		GrantedAbilities:  copyStringSlice(src.GrantedAbilities),
		Power:             src.Power,
		Toughness:         src.Toughness,
		VariableToughness: src.VariableToughness,
		PrintedPTKnown:    src.PrintedPTKnown,
		StartingLoyalty:   src.StartingLoyalty,
		Layout:            src.Layout,
		Faces:             copyFaceSlice(src.Faces),
		ActiveFace:        src.ActiveFace,
	}
}

// Clone returns an independent deep copy. The engine hands
// PrintedValues to card files so they can edit the "except" clause
// into them; every such edit works on a copy of the source's
// values, never on the source card itself.
func (v PrintedValues) Clone() PrintedValues {
	out := v
	out.Colors = copyStringSlice(v.Colors)
	out.ColorIdentity = copyStringSlice(v.ColorIdentity)
	out.ProducedMana = copyStringSlice(v.ProducedMana)
	out.Keywords = copyStringSlice(v.Keywords)
	out.GrantedAbilities = copyStringSlice(v.GrantedAbilities)
	out.Faces = copyFaceSlice(v.Faces)
	return out
}

// --- the "except" clause helpers ------------------------------
//
// Every ETB copy effect in print has an "except" tail — "except
// it's an artifact in addition to its other types", "except its
// name is Sakashima the Impostor", "except it isn't legendary".
// They are all edits to the copiable values before those values
// land, so they are all methods here rather than knobs on the
// engine.

// SetName overrides the copied name (Sakashima the Impostor keeps
// its own). Also rewrites the active face's name so the multi-face
// invariant — flat printed fields equal Faces[ActiveFace] — still
// holds after the copy lands.
func (v *PrintedValues) SetName(name string) {
	v.Name = name
	if v.ActiveFace >= 0 && v.ActiveFace < len(v.Faces) {
		v.Faces[v.ActiveFace].Name = name
	}
}

// AddCardType adds a card type the copy has "in addition to its
// other types" (Phyrexian Metamorph's artifact). Idempotent, and
// inserted in CR 205.1 printed order, so a copied creature reads
// "Artifact Creature — Wurm" the way the real card does rather than
// "Creature Artifact".
func (v *PrintedValues) AddCardType(t string) {
	supers, types, subs := ParseTypeLine(v.TypeLine)
	for _, existing := range types {
		if existing == t {
			return
		}
	}
	at := len(types)
	for i, existing := range types {
		if cardTypeRank(t) < cardTypeRank(existing) {
			at = i
			break
		}
	}
	types = append(types, "")
	copy(types[at+1:], types[at:])
	types[at] = t
	v.setTypeLine(supers, types, subs)
}

// cardTypeRank orders card types the way they are printed (CR
// 205.1a). An unrecognised type sorts last, which keeps a future
// type the engine has not heard of from jumping the queue.
func cardTypeRank(t string) int {
	for i, known := range []string{
		"Artifact", "Battle", "Conspiracy", "Creature", "Enchantment",
		"Instant", "Kindred", "Land", "Phenomenon", "Plane",
		"Planeswalker", "Scheme", "Sorcery", "Tribal", "Vanguard",
	} {
		if known == t {
			return i
		}
	}
	return 99
}

// AddSupertype adds a supertype the copy has in addition to its own
// (Sakashima is "legendary in addition to its other types").
// Idempotent.
func (v *PrintedValues) AddSupertype(s string) {
	supers, types, subs := ParseTypeLine(v.TypeLine)
	for _, existing := range supers {
		if existing == s {
			return
		}
	}
	supers = append(supers, s)
	v.setTypeLine(supers, types, subs)
}

// RemoveSupertype drops a supertype the copy explicitly does not
// have (Spark Double "isn't legendary" — which is the whole reason
// the card is playable in Commander). No-op when absent.
func (v *PrintedValues) RemoveSupertype(s string) {
	supers, types, subs := ParseTypeLine(v.TypeLine)
	kept := make([]string, 0, len(supers))
	for _, existing := range supers {
		if existing != s {
			kept = append(kept, existing)
		}
	}
	if len(kept) == len(supers) {
		return
	}
	v.setTypeLine(kept, types, subs)
}

// AddSubtype adds a subtype the copy has "in addition to its other
// types" — Phantasmal Image's "it's an Illusion in addition to its
// other types" (CR 707.9b). Idempotent, and appended rather than
// inserted: subtypes have no printed order within their class the
// way card types do (CR 205.3), and the added one is the exception
// the card names, so it reads last where a player expects it.
//
// The ADD is deliberately not a SET. "Except it's a 4/4 black
// Zombie" REPLACES the creature types, which is a different clause
// and belongs in the token-copy template rather than here — see
// retypedTypeLine in the catalog's token_copy.go.
func (v *PrintedValues) AddSubtype(s string) {
	supers, types, subs := ParseTypeLine(v.TypeLine)
	for _, existing := range subs {
		if existing == s {
			return
		}
	}
	subs = append(subs, s)
	v.setTypeLine(supers, types, subs)
}

// RemoveSubtype drops a subtype the copy explicitly does not have.
// The mirror of RemoveSupertype, and a no-op when absent.
func (v *PrintedValues) RemoveSubtype(s string) {
	supers, types, subs := ParseTypeLine(v.TypeLine)
	kept := make([]string, 0, len(subs))
	for _, existing := range subs {
		if existing != s {
			kept = append(kept, existing)
		}
	}
	if len(kept) == len(subs) {
		return
	}
	v.setTypeLine(supers, types, kept)
}

// HasSubtype reports whether the copied values carry subtype `s`
// (case-sensitive, Scryfall capitalisation). The subtype twin of
// HasCardType, for an except clause that branches on what was
// copied.
func (v PrintedValues) HasSubtype(s string) bool {
	_, _, subs := ParseTypeLine(v.TypeLine)
	for _, existing := range subs {
		if existing == s {
			return true
		}
	}
	return false
}

// GrantAbility is "except it has '<ability>'" (CR 707.9a) — the
// clause that gives the copy a triggered, static or activated
// ability the copied card never had.
//
// `name` names an ability bundle the catalog registered
// (`effects.Spec.Grants`); only the NAME is stored, which is what
// lets the grant be copied again and survive a snapshot. See
// copy_grants.go for how the name becomes abilities.
//
// Idempotent, and order-preserving: granting the same bundle twice
// is one grant, and two different bundles keep the order the clause
// applied them in, which is the order their abilities are read in.
func (v *PrintedValues) GrantAbility(name string) {
	key := GrantKey(name)
	if key == "" {
		return
	}
	for _, existing := range v.GrantedAbilities {
		if existing == key {
			return
		}
	}
	v.GrantedAbilities = append(v.GrantedAbilities, key)
}

// MakeToken stamps the "Token" supertype on the copiable values —
// what CR 111.13 and CR 608.3f ask for when a copy of a permanent
// spell becomes a token as it resolves.
//
// "Token" is a supertype in this engine's parser (`isSupertype`) and
// `Card.IsToken` is the printed type line, so this one edit is what
// makes the object a token to the CR 704.5d existence check, the
// bounce path and the client. It is PREPENDED rather than appended,
// because that is how a token's type line is printed and how the
// catalog's own `tokenTypeLine` builds one: "Token Legendary
// Creature — Human Advisor". Idempotent, so a copy of a token stays
// one Token, and it goes through setTypeLine so the multi-face
// invariant survives (CR 707.10g — a copy of a double-faced
// permanent spell is double-faced too).
func (v *PrintedValues) MakeToken() {
	supers, types, subs := ParseTypeLine(v.TypeLine)
	for _, existing := range supers {
		if existing == "Token" {
			return
		}
	}
	supers = append([]string{"Token"}, supers...)
	v.setTypeLine(supers, types, subs)
}

// HasCardType reports whether the copied values carry card type `t`
// (case-sensitive, Scryfall capitalisation). Card files branch on
// it for the "if it's a creature / if it's a planeswalker" halves of
// Spark Double's except clause.
func (v PrintedValues) HasCardType(t string) bool {
	_, types, _ := ParseTypeLine(v.TypeLine)
	for _, existing := range types {
		if existing == t {
			return true
		}
	}
	return false
}

// setTypeLine rewrites TypeLine — and the active face's, so ADR
// 0034's invariant survives — from a parsed triple.
func (v *PrintedValues) setTypeLine(supers, types, subs []string) {
	line := composeTypeLine(supers, types, subs)
	v.TypeLine = line
	if v.ActiveFace >= 0 && v.ActiveFace < len(v.Faces) {
		v.Faces[v.ActiveFace].TypeLine = line
	}
}

// composeTypeLine rebuilds a Scryfall-style type line from parsed
// parts: "Legendary Artifact Creature — Human Rogue". The em-dash
// and its surrounding spaces appear only when there are subtypes,
// which is what ParseTypeLine expects to read back.
func composeTypeLine(supertypes, types, subtypes []string) string {
	left := ""
	for _, s := range supertypes {
		if left != "" {
			left += " "
		}
		left += s
	}
	for _, t := range types {
		if left != "" {
			left += " "
		}
		left += t
	}
	if len(subtypes) == 0 {
		return left
	}
	right := ""
	for _, s := range subtypes {
		if right != "" {
			right += " "
		}
		right += s
	}
	if left == "" {
		return right
	}
	return left + " — " + right
}

// --- applying and undoing a copy ------------------------------

// IsCopy reports whether a copy effect is currently applied to this
// permanent.
func (c Card) IsCopy() bool { return c.PrintedSelf != nil }

// applyCopy materialises `v` onto the card, stashing what the card
// was so the battlefield-leave path can put it back (CR 400.7).
//
// `src` is the card the values came from, and is read for the two
// things PrintedValues deliberately cannot carry: the card-carried
// mana and activated abilities, which exist only on TOKENS (a token
// has no oracle ID for the catalog to key on, so copying a Treasure
// token means copying its ability slice or copying nothing).
//
// Structurally the twin of SetFace (face.go): both take a set of
// printed values and write them into the flat printed fields, and
// both leave every downstream reader — Is*() predicates, the layer
// engine's printed baseline, CatalogKey, the wire projection —
// correct without knowing anything happened.
//
// Idempotent in the sense that matters: a permanent that is already
// a copy and becomes a copy of something else keeps its ORIGINAL
// PrintedSelf, so it still reverts to the right card on death.
func (c *Card) applyCopy(v PrintedValues, src Card) {
	if c.PrintedSelf == nil {
		own := CopiableValuesOf(*c)
		c.PrintedSelf = &own
	}
	c.setPrintedValues(v)
	c.ManaAbilities = append([]ManaAbilityShape(nil), src.ManaAbilities...)
	c.ActivatedAbilities = append([]ActivatedAbilityShape(nil), src.ActivatedAbilities...)
}

// setPrintedValues writes `v` into the flat printed fields and
// nothing else — no PrintedSelf stash, no card-carried ability
// slices. applyCopy is this plus the two things a PERMANENT copy
// needs; the spell-copy path (spell_copy.go) uses it bare, because a
// copy of a spell has no printed self to revert to (it is not a card
// and never leaves the stack for a zone), and so does the token a
// resolving copy of a permanent spell becomes (CR 608.3f).
func (c *Card) setPrintedValues(v PrintedValues) {
	v = v.Clone()
	c.OracleID = v.OracleID
	c.TokenKey = v.TokenKey
	c.ScryfallID = v.ScryfallID
	c.Name = v.Name
	c.TypeLine = v.TypeLine
	c.ManaCost = v.ManaCost
	c.Colors = v.Colors
	c.ColorIdentity = v.ColorIdentity
	c.ProducedMana = v.ProducedMana
	c.Keywords = v.Keywords
	c.GrantedAbilities = v.GrantedAbilities
	c.Power = v.Power
	c.Toughness = v.Toughness
	c.VariableToughness = v.VariableToughness
	c.PrintedPTKnown = v.PrintedPTKnown
	c.StartingLoyalty = v.StartingLoyalty
	c.Layout = v.Layout
	c.Faces = v.Faces
	c.ActiveFace = v.ActiveFace
	// The printed baseline just changed underneath the layer
	// engine's cache. Callers are on a battlefield-entry path that
	// bumps layerVersion via the zone-move listener, but clearing
	// here means the card is never observable as "copied values with
	// a pre-copy effective cache" even for one read.
	c.effective = nil
}

// restorePrintedSelf undoes a copy effect, putting the card's own
// printed values back. Called from the battlefield-leave path; a
// no-op for the ~everything that never was a copy.
//
// The card-carried ability slices are cleared rather than restored:
// they are catalog-rebuildable data (see the drift plan's `rebuilt`
// disposition for them), the restored oracle ID is what rebuilds
// them, and off the battlefield nothing reads them at all.
func (c *Card) restorePrintedSelf() {
	if c.PrintedSelf == nil {
		return
	}
	v := c.PrintedSelf.Clone()
	c.OracleID = v.OracleID
	c.TokenKey = v.TokenKey
	c.ScryfallID = v.ScryfallID
	c.Name = v.Name
	c.TypeLine = v.TypeLine
	c.ManaCost = v.ManaCost
	c.Colors = v.Colors
	c.ColorIdentity = v.ColorIdentity
	c.ProducedMana = v.ProducedMana
	c.Keywords = v.Keywords
	// CR 400.7 again: the grant applied to the PERMANENT, so the card
	// that goes to the graveyard is not an Illusion and has no
	// sacrifice trigger. PrintedSelf carries the card's own (almost
	// always empty) grant list, so this restores rather than clears.
	c.GrantedAbilities = v.GrantedAbilities
	c.Power = v.Power
	c.Toughness = v.Toughness
	c.VariableToughness = v.VariableToughness
	c.PrintedPTKnown = v.PrintedPTKnown
	c.StartingLoyalty = v.StartingLoyalty
	c.Layout = v.Layout
	c.Faces = v.Faces
	c.ActiveFace = v.ActiveFace
	c.ManaAbilities = nil
	c.ActivatedAbilities = nil
	c.PrintedSelf = nil
	c.effective = nil
}

// copyStringSlice deep-copies a string slice, preserving nil.
func copyStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	return append([]string(nil), in...)
}

// copyFaceSlice deep-copies a face list including each face's own
// Colors slice. `copyFaces` in snapshot.go does the same job for the
// snapshot mirror; this one is the domain-side twin, kept here so
// copy.go reads without a jump into the persistence file.
func copyFaceSlice(in []Face) []Face {
	if len(in) == 0 {
		return nil
	}
	out := make([]Face, len(in))
	copy(out, in)
	for i := range out {
		out[i].Colors = copyStringSlice(in[i].Colors)
	}
	return out
}
