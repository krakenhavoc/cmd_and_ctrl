package game

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// text_words.go — changing one word of an object's text (CR 612.2;
// ADR 0160's 2026-10-15 amendment, #2779).
//
// Sleight of Mind: "Change the text of target spell or permanent by
// replacing all instances of one color word with another." Magical
// Hack does the same for a basic land type, Artificial Evolution for a
// creature type.
//
// # There is no text to rewrite
//
// A card's rules are Go closures and data on its catalog entry, so the
// words are not text either. They are arguments: OfColor("B") inside a
// target clause, "protection from black" in a keyword list, "Swamp" in
// a type line, "Goblin" in a lord's TribeFilter. So a change is not a
// rewrite. It is a per-object substitution, TextWords, that every
// reader of a word consults through the object whose text it is:
//
//   - The object's own TYPE LINE and printed KEYWORDS are translated as
//     the layer pass builds its baseline (applyToBaseline), so a Forest
//     changed to an Island taps for blue (CR 305.6, the card's ruling)
//     and swampwalk becomes islandwalk. Granted keywords arrive in layer
//     6 and are never translated (CR 612.3).
//   - A TARGET clause reads its words through the source of the spell or
//     ability: the targeting walk lends the candidate the source's
//     TextWords (Card.wordLens), and the effects package's colour and
//     subtype predicates read HasColorWord / HasSubtypeWord, which look
//     through it. Doom Blade's "nonblack" changed to "nonblue" is then
//     the legal set, the announce check and the resolution re-check
//     (CR 608.2b) at once.
//   - A STATIC that holds its source reads TextColor / TextSubtype off
//     it (a lord's tribes, a colour anthem).
//
// What a closure hard-codes without going through one of those doors is
// invisible to the substitution; ADR 0160's amendment lists the classes.
//
// # The record
//
// One ScopedEffect per change carrying one ModChangeText mod, pinned to
// the changed object: a permanent (PinnedTo) or a spell on the stack
// (ADR 0104's stack pin, which a permanent spell hands to the permanent
// it becomes, CR 400.7a — "if you change the text of a spell which is
// to become a permanent, the permanent will retain the text change").
// materialiseTextWordsLocked composes every record on an object in
// timestamp order (CR 613.7) onto Card.TextWords before the layer pass,
// beside the text boxes.

// ModChangeText is "replace every instance of one word in <affected>'s
// text with another" (layer 3, CR 612.2). Reads Words, which holds
// exactly ONE change. Applied before the pass like ModExchangeText, so
// its reader is readerText.
const ModChangeText ModKind = "changeText"

// TextWordKind is the class of word a change replaces. CR 612.2 limits
// a text change to words used in the right way: a colour word as a
// colour word, a land type as a land type.
type TextWordKind string

const (
	// TextWordColor is a colour word, stored as its WUBRG letter.
	TextWordColor TextWordKind = "color"
	// TextWordBasicLandType is one of CR 305.6's five.
	TextWordBasicLandType TextWordKind = "basicLandType"
	// TextWordCreatureType is a CR 205.3m creature type.
	TextWordCreatureType TextWordKind = "creatureType"
)

// TextWordChange is one change: every instance of From becomes To.
// Plain data, on a ScopedEffect's mod.
type TextWordChange struct {
	Kind TextWordKind `json:"kind"`
	From string       `json:"from"`
	To   string       `json:"to"`
}

// ChangeTextMod is "replace From with To in <affected>'s text".
func ChangeTextMod(ch TextWordChange) Mod {
	return Mod{Kind: ModChangeText, Words: []TextWordChange{ch}}
}

// canonicalTextWord returns `w` in the spelling a kind stores, and
// false when it is not a word of that kind.
func canonicalTextWord(kind TextWordKind, w string) (string, bool) {
	switch kind {
	case TextWordColor:
		switch strings.ToUpper(strings.TrimSpace(w)) {
		case "W", "WHITE":
			return "W", true
		case "U", "BLUE":
			return "U", true
		case "B", "BLACK":
			return "B", true
		case "R", "RED":
			return "R", true
		case "G", "GREEN":
			return "G", true
		}
		return "", false
	case TextWordBasicLandType:
		if !IsBasicLandType(w) {
			return "", false
		}
		return canonicalBasicLandType(w), true
	case TextWordCreatureType:
		return CanonicalCreatureType(w)
	}
	return "", false
}

// Normalised returns the change in canonical spelling, and an error
// when it is not one CR 612.2 allows: an unknown kind, a word that is
// not of its kind, or a word replaced by itself ("It can't change a
// word to the same word", the cards' ruling).
func (ch TextWordChange) Normalised() (TextWordChange, error) {
	from, ok := canonicalTextWord(ch.Kind, ch.From)
	if !ok {
		return ch, fmt.Errorf("%q is not a %s word", ch.From, ch.Kind)
	}
	to, ok := canonicalTextWord(ch.Kind, ch.To)
	if !ok {
		return ch, fmt.Errorf("%q is not a %s word", ch.To, ch.Kind)
	}
	if from == to {
		return ch, fmt.Errorf("a text change replaces %q with itself", from)
	}
	return TextWordChange{Kind: ch.Kind, From: from, To: to}, nil
}

// textWordsModProblem is the registration and restore check: exactly one
// sound change on a changeText mod, and none on any other kind. "" when
// sound.
func textWordsModProblem(m Mod) string {
	if m.Kind == ModChangeText {
		if len(m.Words) != 1 {
			return fmt.Sprintf("a changeText mod carries exactly one change, got %d", len(m.Words))
		}
		norm, err := m.Words[0].Normalised()
		if err != nil {
			return err.Error()
		}
		if norm != m.Words[0] {
			return fmt.Sprintf("a changeText mod's change %+v is not canonical", m.Words[0])
		}
		return ""
	}
	if len(m.Words) != 0 {
		return fmt.Sprintf("mod %q carries a text change, which only changeText reads", m.Kind)
	}
	return ""
}

// TextWords is the net substitution an object's text is under: for each
// kind, the word a printed word now reads as. A word with no entry reads
// as itself. IMMUTABLE once built — Card.TextWords points at one and
// Clone shares the pointer — so it is replaced, never written through.
type TextWords struct {
	words map[TextWordKind]map[string]string
	key   string
}

// with returns the substitution after one more change, applied on top
// (CR 613.7: a later change rewrites the text the earlier ones left).
// Every word that reads From now reads To; that includes words an
// earlier change turned INTO From, and From itself when nothing has
// changed it yet. A word that ends up reading as itself drops out.
func (w *TextWords) with(ch TextWordChange) *TextWords {
	out := &TextWords{words: map[TextWordKind]map[string]string{}}
	if w != nil {
		for k, m := range w.words {
			cp := make(map[string]string, len(m))
			for from, to := range m {
				cp[from] = to
			}
			out.words[k] = cp
		}
	}
	m := out.words[ch.Kind]
	if m == nil {
		m = map[string]string{}
		out.words[ch.Kind] = m
	}
	for printed, now := range m {
		if now == ch.From {
			m[printed] = ch.To
		}
	}
	if _, changed := m[ch.From]; !changed {
		m[ch.From] = ch.To
	}
	for printed, now := range m {
		if printed == now {
			delete(m, printed)
		}
	}
	if len(m) == 0 {
		delete(out.words, ch.Kind)
	}
	out.key = out.describe()
	return out
}

// describe is a canonical rendering, used both as the equality key and
// for the wire: "black → blue", in kind order then word order.
func (w *TextWords) describe() string {
	return strings.Join(w.Lines(), "; ")
}

// Lines renders each changed word as "<printed> → <now>", colours as
// their English words: the wire's CardView.text_changes.
func (w *TextWords) Lines() []string {
	if w == nil {
		return nil
	}
	var out []string
	for _, k := range []TextWordKind{TextWordColor, TextWordBasicLandType, TextWordCreatureType} {
		m := w.words[k]
		printed := make([]string, 0, len(m))
		for p := range m {
			printed = append(printed, p)
		}
		sort.Slice(printed, func(i, j int) bool { return wordOrder(k, printed[i]) < wordOrder(k, printed[j]) })
		for _, p := range printed {
			out = append(out, textWordName(k, p)+" → "+textWordName(k, m[p]))
		}
	}
	return out
}

// wordOrder sorts colours WUBRG and land types in CR 305.6 order, and
// anything else by name.
func wordOrder(k TextWordKind, w string) string {
	switch k {
	case TextWordColor:
		return string(rune('0' + strings.Index("WUBRG", w)))
	case TextWordBasicLandType:
		for i, b := range BasicLandTypes {
			if b == w {
				return string(rune('0' + i))
			}
		}
	}
	return w
}

func textWordName(k TextWordKind, w string) string {
	if k == TextWordColor {
		return ColorName(w)
	}
	return w
}

func (w *TextWords) equal(o *TextWords) bool {
	if w == nil || o == nil {
		return w == o
	}
	return w.key == o.key
}

// word is what the printed word `w` of kind `k` reads as now.
func (w *TextWords) word(k TextWordKind, printed string) string {
	if w == nil {
		return printed
	}
	m := w.words[k]
	if len(m) == 0 {
		return printed
	}
	canon, ok := canonicalTextWord(k, printed)
	if !ok {
		return printed
	}
	if now, ok := m[canon]; ok {
		return now
	}
	return printed
}

// color is the colour letter the colour word `letter` reads as.
func (w *TextWords) color(letter string) string { return w.word(TextWordColor, letter) }

// subtype is what the subtype word `sub` reads as: through the land-type
// map for one of the five basic land types, the creature-type map for a
// creature type, and unchanged for anything else ("Equipment").
func (w *TextWords) subtype(sub string) string {
	if w == nil {
		return sub
	}
	if IsBasicLandType(sub) {
		return w.word(TextWordBasicLandType, sub)
	}
	return w.word(TextWordCreatureType, sub)
}

// keyword is a printed keyword token as the text now reads: the colour
// in "protection from <colour>" and "hexproof from <colour>", the
// creature type in "protection from <creature type>" (read and respelled
// through the protection parser, protectionFromCreatureType, #3872), and
// the land type in a basic landwalk ("swampwalk"). Every other token —
// and these shapes for a word the substitution leaves alone — comes back
// unchanged.
func (w *TextWords) keyword(tok string) string {
	if w == nil {
		return tok
	}
	lower := strings.ToLower(tok)
	for _, prefix := range []string{protectionPrefix, hexproofFromPrefix} {
		if !strings.HasPrefix(lower, prefix) {
			continue
		}
		rest := strings.TrimSpace(lower[len(prefix):])
		letter, ok := protectionColors[rest]
		if !ok {
			if prefix == protectionPrefix {
				return w.protectionFromCreatureType(tok)
			}
			return tok
		}
		now := w.color(letter)
		if now == letter {
			return tok
		}
		return tok[:len(prefix)] + ColorName(now)
	}
	if strings.HasSuffix(lower, "walk") {
		land := lower[:len(lower)-len("walk")]
		if !IsBasicLandType(land) {
			return tok
		}
		now := w.word(TextWordBasicLandType, land)
		if equalFoldASCII(now, land) {
			return tok
		}
		return strings.ToLower(now) + "walk"
	}
	return tok
}

// protectionFromCreatureType translates the creature type of a
// "protection from <creature type, plural>" token ("protection from
// Goblins" under "Goblin → Elf" is "protection from Elves"), and returns
// any other token unchanged.
//
// The token names its type in the plural and the substitution holds
// singulars, so the type is read the way the protection parser reads it
// (ParseProtectionQuality, through singulars and CanonicalCreatureType),
// and the new word is spelled back as a plural that the same parser
// reads as the new type. A spelling that would not round-trip is never
// written: the token then stays as printed rather than becoming a
// protection from nothing.
func (w *TextWords) protectionFromCreatureType(tok string) string {
	q, ok := ParseProtectionQuality(tok)
	if !ok || q.Kind != ProtectionQualitySubtype {
		return tok
	}
	now := w.word(TextWordCreatureType, q.Value)
	if now == q.Value {
		return tok
	}
	for _, spelled := range creatureTypePlurals(now) {
		if back, ok := parseQuality(spelled); ok && back.Kind == ProtectionQualitySubtype && back.Value == now {
			return tok[:len(protectionPrefix)] + spelled
		}
	}
	return tok
}

// creatureTypePlurals are the plural spellings to try for a creature
// type, most natural first: Elf → Elves, Goblin → Goblins, and a word
// English leaves alone or that ends in a sibilant (Merfolk, Fox) as it is.
func creatureTypePlurals(t string) []string {
	lower := strings.ToLower(t)
	var out []string
	switch {
	case strings.HasSuffix(lower, "f"):
		out = append(out, t[:len(t)-1]+"ves")
	case strings.HasSuffix(lower, "folk"), strings.HasSuffix(lower, "kin"),
		strings.HasSuffix(lower, "s"), strings.HasSuffix(lower, "x"), strings.HasSuffix(lower, "z"),
		strings.HasSuffix(lower, "ch"), strings.HasSuffix(lower, "sh"):
		out = append(out, t)
	}
	return append(out, t+"s", t)
}

// keywords translates a keyword list, returning the input itself when
// nothing in it changes (it is never written through).
func (w *TextWords) keywords(list []string) []string {
	if w == nil {
		return list
	}
	var out []string
	for i, tok := range list {
		now := w.keyword(tok)
		if out == nil && now == tok {
			continue
		}
		if out == nil {
			out = append(make([]string, 0, len(list)), list[:i]...)
		}
		out = append(out, now)
	}
	if out == nil {
		return list
	}
	return out
}

// applyToBaseline is layer 3 for this effect class on the layer pass's
// baseline: the type line's basic land types and creature types and the
// printed keywords, as the text now reads. Runs on the fresh baseline,
// before layer 4 and layer 6 apply anything, so a type-changing or
// ability-granting effect is never translated (CR 612.3).
func (w *TextWords) applyToBaseline(ch *Characteristic) {
	if w == nil {
		return
	}
	var subs []string
	for i, sub := range ch.Subtypes {
		now := w.subtype(sub)
		if subs == nil && now == sub {
			continue
		}
		if subs == nil {
			subs = append(make([]string, 0, len(ch.Subtypes)), ch.Subtypes[:i]...)
		}
		// Swamp → Island on an Island Swamp leaves one Island.
		if !typeListHas(subs, now) {
			subs = append(subs, now)
		}
	}
	if subs != nil {
		ch.Subtypes = subs
	}
	ch.Abilities = w.keywords(ch.Abilities)
}

// --- readers -------------------------------------------------------

// TextColor is the colour letter the colour word `letter` reads as in
// this object's text. For a static or trigger that holds its source:
// "green creatures you control" is target.HasColor(source.TextColor("G")).
func (c *Card) TextColor(letter string) string {
	if c == nil {
		return letter
	}
	return c.TextWords.color(letter)
}

// TextSubtype is the subtype `sub` reads as in this object's text — a
// basic land type or a creature type; any other subtype unchanged.
func (c *Card) TextSubtype(sub string) string {
	if c == nil {
		return sub
	}
	return c.TextWords.subtype(sub)
}

// TextColors is TextColor over a list of colour words, each colour once
// and in the order first read: a colour-setting effect's printed colours
// ("becomes a white and blue Bird") as this object's text now says them
// (#3984). Under "white → blue" that is one colour, blue.
func (c *Card) TextColors(letters ...string) []string {
	return textWordsOnce(letters, c.TextColor)
}

// TextSubtypes is TextSubtype over a list of subtype words, each once: a
// type-setting effect's printed types ("becomes a Bird", "is a Zombie in
// addition to its other types") as this object's text now says them.
func (c *Card) TextSubtypes(subs ...string) []string {
	return textWordsOnce(subs, c.TextSubtype)
}

// textWordsOnce maps each word through `read`, dropping a repeat.
func textWordsOnce(words []string, read func(string) string) []string {
	if len(words) == 0 {
		return words
	}
	out := make([]string, 0, len(words))
	for _, w := range words {
		now := read(w)
		if !typeListHas(out, now) {
			out = append(out, now)
		}
	}
	return out
}

// TextKeyword is the keyword token `tok` as this object's text now
// reads it: the colour of "protection from <colour>", the land type of
// a basic landwalk. For a static that grants a keyword its source
// prints ("other Goblins have mountainwalk").
func (c *Card) TextKeyword(tok string) string {
	if c == nil {
		return tok
	}
	return c.TextWords.keyword(tok)
}

// TokenThroughText is the token a description in this object's text
// makes, as the text now says it (CR 612.2, #4005): "create a 2/2 black
// Zombie creature token" under a Sleight of Mind (black → blue) makes a
// blue Zombie, and under an Artificial Evolution (Zombie → Elf) a black
// Elf. The template's colours, its subtypes (a creature type or a basic
// land type), and the words its keywords name ("protection from white",
// "islandwalk") are read through the text. A token the description names
// only by its subtypes is named after them (CR 111.4), so its name
// follows; a token with a name of its own ("named Cadet") keeps it.
//
// Only for a token DESCRIBED by the text. A copy's template is copiable
// values (CR 707.2), which a text change never touches, and a token a
// keyword action defines (living weapon's Germ, incubate's Incubator) is
// no word of the text at all. A template with faces is returned as it is.
// A nil reader, or one whose words are printed, returns `tmpl` itself.
func (c *Card) TokenThroughText(tmpl Card) Card {
	if c == nil || c.TextWords == nil || len(tmpl.Faces) > 0 {
		return tmpl
	}
	w := c.TextWords
	if len(tmpl.Colors) > 0 {
		tmpl.Colors = textWordsOnce(tmpl.Colors, w.color)
	}
	supers, types, subs := ParseTypeLine(tmpl.TypeLine)
	if now := textWordsOnce(subs, w.subtype); !slices.Equal(now, subs) {
		if tmpl.Name == strings.Join(subs, " ") {
			tmpl.Name = strings.Join(now, " ")
		}
		tmpl.TypeLine = composeTypeLine(supers, types, now)
	}
	tmpl.Keywords = w.keywords(tmpl.Keywords)
	return tmpl
}

// ThroughTextOf is f with the creature types it names read as `reader`'s
// text now says them (CR 612.2, #4005): Gisa and Geralf's "a Zombie
// creature spell" under an Artificial Evolution (Zombie → Elf) opens Elf
// creature spells. A type chosen as the source entered (FromChosenType,
// Realmwalker) is not a word of its text and is left alone; so is a
// filter whose words nothing changed, which comes back as it was.
func (f PermissionFilter) ThroughTextOf(reader *Card) PermissionFilter {
	if reader == nil || reader.TextWords == nil {
		return f
	}
	if f.CreatureType != "" && !f.FromChosenType {
		f.CreatureType = reader.TextSubtype(f.CreatureType)
	}
	for i, t := range f.CreatureTypesAny {
		if t != "" {
			f.CreatureTypesAny[i] = reader.TextSubtype(t)
		}
	}
	return f
}

// TextChanges renders this object's changed words for the wire
// ("black → blue"), nil when none are.
func (c *Card) TextChanges() []string { return c.TextWords.Lines() }

// ThroughTextOf lends `c` the text of `reader`: a predicate that names a
// colour or subtype word and is asked about `c` then reads that word as
// `reader`'s text now says it (HasColorWord, HasSubtypeWord). The
// targeting walk does this with the spell or ability's source; an
// effect that judges cards against its own clause does it with its
// source. A nil reader lends nothing.
//
// The lens lives on the VALUE copy returned and nowhere else: never
// store the result in a zone.
func ThroughTextOf(reader *Card, c Card) Card {
	if reader == nil {
		c.wordLens = nil
		return c
	}
	c.wordLens = reader.TextWords
	return c
}

// HasColorWord is HasColor for a colour WORD in some object's text: the
// colour `letter` reads as through the lens ThroughTextOf lent this
// copy, or `letter` itself when it was lent none.
func (c Card) HasColorWord(letter string) bool {
	return c.HasColor(c.wordLens.color(letter))
}

// HasSubtypeWord is HasSubtype for a subtype WORD in some object's text
// (a basic land type or a creature type), read through the lens.
func (c Card) HasSubtypeWord(sub string) bool {
	return c.HasSubtype(c.wordLens.subtype(sub))
}

// ThroughTextOf is q with its subtype and colour words read as `reader`'s
// text now says them (CR 612.2): Seasinger's "When you control no
// Islands" under a Magical Hack asks about Forests. q itself is never
// written through; with nothing to translate it comes back as it was.
func (q PermanentQuery) ThroughTextOf(reader *Card) PermanentQuery {
	if reader == nil || reader.TextWords == nil {
		return q
	}
	if len(q.Subtypes) > 0 {
		subs := make([]string, len(q.Subtypes))
		for i, s := range q.Subtypes {
			subs[i] = reader.TextSubtype(s)
		}
		q.Subtypes = subs
	}
	if len(q.Colors) > 0 {
		colors := make([]string, len(q.Colors))
		for i, c := range q.Colors {
			colors[i] = reader.TextColor(c)
		}
		q.Colors = colors
	}
	return q
}

// PermanentQueriesThroughTextOf is ThroughTextOf over an any-of list.
func PermanentQueriesThroughTextOf(qs []PermanentQuery, reader *Card) []PermanentQuery {
	if reader == nil || reader.TextWords == nil || len(qs) == 0 {
		return qs
	}
	out := make([]PermanentQuery, len(qs))
	for i, q := range qs {
		out[i] = q.ThroughTextOf(reader)
	}
	return out
}

// LentColor is the colour letter the colour word `letter` reads as in
// the text ThroughTextOf lent this copy, or `letter` when it was lent
// none. For a target clause whose word is not about the candidate
// itself: "target creature whose controller controls a black permanent".
func (c Card) LentColor(letter string) string { return c.wordLens.color(letter) }

// LentSubtype is LentColor for a subtype word: Seasinger's "whose
// controller controls an Island", Vedalken Shackles' "the number of
// Islands you control".
func (c Card) LentSubtype(sub string) string { return c.wordLens.subtype(sub) }

// manaTags reads a printed mana restriction's words through this text
// (#3984): a "subtype:" tag's subtype and a "color:" tag's colour, each
// the word as the text now says it. Every other tag, and a compound one,
// is a word no text change names and comes back unchanged. Writes `tags`
// in place, so callers pass a copy.
func (w *TextWords) manaTags(tags []string) []string {
	if w == nil {
		return tags
	}
	for i, t := range tags {
		switch {
		case strings.HasPrefix(t, manaTagSubtype):
			tags[i] = manaTagSubtype + w.subtype(t[len(manaTagSubtype):])
		case strings.HasPrefix(t, manaTagColor):
			tags[i] = manaTagColor + w.color(t[len(manaTagColor):])
		}
	}
	return tags
}

// The two mana restriction tag prefixes that name a word of a card's
// text (ManaRestrictSubtype, ManaRestrictColor).
const (
	manaTagSubtype = "subtype:"
	manaTagColor   = "color:"
)

// --- materialising -------------------------------------------------

// textWordsFor composes the changeText records on `c` in timestamp
// order (ties keep registration order), and nil when none applies.
func textWordsFor(c *Card, onStack bool, records []*ScopedEffect) *TextWords {
	var mine []*ScopedEffect
	for _, e := range records {
		if onStack {
			if !pinsStackObject(e.Affected, c.InstanceID, c.ObjectEpoch) {
				continue
			}
		} else if !affectsObject(e.Affected, c) {
			continue
		}
		mine = append(mine, e)
	}
	if len(mine) == 0 {
		return nil
	}
	sort.SliceStable(mine, func(i, j int) bool { return mine[i].Timestamp < mine[j].Timestamp })
	var w *TextWords
	for _, e := range mine {
		for _, m := range e.Mods {
			if m.Kind == ModChangeText && len(m.Words) == 1 {
				w = w.with(m.Words[0])
			}
		}
	}
	if w == nil || len(w.words) == 0 {
		return nil
	}
	return w
}

// materialiseTextWordsLocked brings every permanent's and every spell's
// Card.TextWords in line with the changeText records. Idempotent, and
// bumps the layer version only when it writes, as the text boxes do.
//
// Caller must hold g.mu (write).
func (g *Game) materialiseTextWordsLocked() {
	var records []*ScopedEffect
	for i := range g.ScopedEffects {
		if scopedEffectHasMod(&g.ScopedEffects[i], ModChangeText) {
			records = append(records, &g.ScopedEffects[i])
		}
	}
	settle := func(z *Zone, onStack bool) {
		if z == nil {
			return
		}
		for i := range z.Cards {
			c := &z.Cards[i]
			if len(records) == 0 && c.TextWords == nil {
				continue
			}
			want := textWordsFor(c, onStack, records)
			if c.TextWords.equal(want) {
				continue
			}
			c.TextWords = want
			g.layerVersion.Add(1)
		}
	}
	settle(g.Battlefield, false)
	settle(g.Stack, true)
}

// ChangeTextWordForEffect replaces every instance of one word with
// another in the text of `object` — a permanent, or a spell on the stack
// — for duration `d` (CR 612.2). Glamerdye, Sleight of Mind, Magical
// Hack, Mind Bend. The change is normalised first, and refused when it
// is not one CR 612.2 allows (Normalised).
//
// "It can target a card with no appropriate words on it": the record is
// registered whatever the object says, because what its text says is
// not something the engine can read. A record on a permanent spell goes
// with it to the battlefield (CR 400.7a). Reports whether a record was
// registered.
//
// Caller must hold g.mu (write); effects call it from the resolution
// frame, which does.
func (g *Game) ChangeTextWordForEffect(sourceID, object uuid.UUID, change TextWordChange, d Duration, label string) bool {
	norm, err := change.Normalised()
	if err != nil || object == uuid.Nil {
		return false
	}
	var (
		affected []AffectedObject
		dur      Duration
	)
	if _, ok := g.battlefieldCardLocked(object); ok {
		affected = g.PinnedObjectsLocked(object)
		dur = g.PinnedTo(d, object)
	} else if c, ok := g.stackCardLocked(object); ok {
		if item := g.StackMeta[object]; item == nil || item.Kind != StackItemSpell {
			return false
		}
		affected = []AffectedObject{PinStackObject(object, c.ObjectEpoch)}
		dur = g.PinnedToStack(d, object)
	}
	if len(affected) == 0 {
		return false
	}
	if !g.registerScopedEffectLocked(sourceID, affected, []Mod{ChangeTextMod(norm)}, dur, label, timeNowUnixNano()) {
		return false
	}
	g.materialiseTextWordsLocked()
	g.RecomputeLayersIfStaleLocked()
	return true
}

func cloneTextWordChanges(in []TextWordChange) []TextWordChange {
	if len(in) == 0 {
		return nil
	}
	return append([]TextWordChange(nil), in...)
}

// --- last-known information ---------------------------------------

// entries is the net substitution as plain data, one change per printed
// word that now reads differently (From is the printed word, To what it
// reads as), in Lines' order. Nil for no change. This is the form the
// last-known records carry (PermanentInfo.TextWords,
// triggerIdentityLKI.TextWords), so a source that has left the
// battlefield is still read with the words it last had (CR 608.2h,
// CR 603.10a; #3872).
func (w *TextWords) entries() []TextWordChange {
	if w == nil || len(w.words) == 0 {
		return nil
	}
	var out []TextWordChange
	for _, k := range []TextWordKind{TextWordColor, TextWordBasicLandType, TextWordCreatureType} {
		m := w.words[k]
		printed := make([]string, 0, len(m))
		for p := range m {
			printed = append(printed, p)
		}
		sort.Slice(printed, func(i, j int) bool { return wordOrder(k, printed[i]) < wordOrder(k, printed[j]) })
		for _, p := range printed {
			out = append(out, TextWordChange{Kind: k, From: p, To: m[p]})
		}
	}
	return out
}

// textWordsFromEntries rebuilds the substitution entries wrote. Each
// entry is a printed word and what it reads as, so they are set side by
// side, never composed (a swap, black → blue and blue → black, is two
// entries, which composing would collapse). An entry that is not a
// sound change of its kind is skipped, so a damaged record reads as the
// printed words rather than as a wrong one. Nil for none.
func textWordsFromEntries(entries []TextWordChange) *TextWords {
	if len(entries) == 0 {
		return nil
	}
	out := &TextWords{words: map[TextWordKind]map[string]string{}}
	for _, e := range entries {
		norm, err := e.Normalised()
		if err != nil {
			continue
		}
		m := out.words[norm.Kind]
		if m == nil {
			m = map[string]string{}
			out.words[norm.Kind] = m
		}
		m[norm.From] = norm.To
	}
	if len(out.words) == 0 {
		return nil
	}
	out.key = out.describe()
	return out
}

// textReaderOf is a stand-in object carrying only a text substitution, for
// a reader whose object has left: Card.TextColor, TextSubtype and
// ThroughTextOf read nothing else. Nil when nothing was changed.
func textReaderOf(w *TextWords) *Card {
	if w == nil {
		return nil
	}
	return &Card{TextWords: w}
}

// TextReaderForEffect is the object whose text a resolving item's words
// are in, as the word readers take it (Card.TextColor, TextSubtype,
// ThroughTextOf), or nil when that text is printed (CR 612.2):
//
//   - a spell: the spell on the stack, read now;
//   - an ability whose source is still the same object on the
//     battlefield: that permanent, read now;
//   - an ability whose source has left: its words as it last existed on
//     the battlefield (CR 608.2h, #3872), from the last-known record.
//     A fetchland changed from "Island" to "Forest" and sacrificed to
//     its own ability searches for a Forest; a card in a graveyard is a
//     new object whose text is printed (CR 400.7), so it is never read.
//
// The answer is a stand-in carrying only the text change, never a live
// card: hold it as long as you like.
//
// Caller must hold g.mu.
func (g *Game) TextReaderForEffect(item *StackItem) *Card {
	if item == nil || item.SourceCardID == uuid.Nil {
		return nil
	}
	if item.Kind != StackItemSpell {
		if rec, ok := g.departedAbilitySourceLocked(item); ok {
			return rec.TextReader()
		}
	}
	if c := findBattlefieldCard(g, item.SourceCardID); c != nil {
		return textReaderOf(c.TextWords)
	}
	if c, ok := g.stackCardLocked(item.SourceCardID); ok {
		return textReaderOf(c.TextWords)
	}
	// Neither a permanent nor a spell: the card is in another zone, where
	// its own text is printed (CR 400.7). An effect still finishing for
	// the permanent it was — a resolution continuation, which carries
	// the card and not the object (OptionPickThen) — reads that
	// permanent's last words while the card has not moved on since it
	// left (departedDamageSourceLocked's rule): Liliana of the Dark
	// Realms' −3 that took her last loyalty still counts the land type
	// her text last named.
	if rec, ok := g.departedDamageSourceLocked(item.SourceCardID, nil); ok {
		return rec.TextReader()
	}
	return nil
}

// TextReaderOfForEffect is TextReaderForEffect for a clause read off an
// object by its id rather than off a stack item — a mana ability's
// condition or amount — read through the card `id` while it is a
// permanent or a spell, and nil otherwise (a card in any other zone has
// its printed text, CR 400.7) or when its words are printed.
//
// Caller must hold g.mu.
func (g *Game) TextReaderOfForEffect(id uuid.UUID) *Card {
	if id == uuid.Nil {
		return nil
	}
	if c := findBattlefieldCard(g, id); c != nil {
		return textReaderOf(c.TextWords)
	}
	if c, ok := g.stackCardLocked(id); ok {
		return textReaderOf(c.TextWords)
	}
	return nil
}
