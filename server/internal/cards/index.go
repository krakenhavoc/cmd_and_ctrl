// Package cards provides the card-metadata layer: a read-only index
// loaded from the Scryfall default-cards bulk dump, plus a disk-
// backed lazy image cache that serves card images via Scryfall's CDN.
//
// The bulk dump lives at $CMDCTRL_DATA_DIR/scryfall/default-cards.json
// and is refreshed weekly by scripts/scryfall-refresh.sh. The file is
// streamed on load (it's ~450MB uncompressed) and only the fields the
// server cares about are kept in memory.
package cards

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Card is the trimmed subset of Scryfall's card record we keep
// in-memory. The S04 set (name + image URIs) grew in S05 to include
// the fields needed for deck validation: type_line, color_identity,
// oracle_text, and the Commander entry from the legalities map.
// Full rules-engine fields (mana cost breakdown, loyalty, power/
// toughness parsing) arrive with the S13+ engine work.
type Card struct {
	ID uuid.UUID `json:"id"`
	// OracleID is the Scryfall oracle-level card identity, stable
	// across printings of the same card (e.g. every printing of
	// Lightning Bolt shares one oracle_id). Used by the S14+ card-
	// effect catalog as the spec key, so a deck importing a set-
	// specific printing still matches the catalog entry. Added in
	// S14 sub-PR 4.
	OracleID    uuid.UUID `json:"oracle_id"`
	Name        string    `json:"name"`
	SetCode     string    `json:"set"`
	SetType     string    `json:"set_type"`
	CollectorNo string    `json:"collector_number"`
	Lang        string    `json:"lang"`
	// Layout is Scryfall's printing layout: "normal", "split",
	// "transform", "token", "art_series", "double_faced_token", etc.
	// Carried because the byName index has to demote token /
	// art-series records when they collide with a real printing.
	Layout string `json:"layout"`
	// TypeLine is Scryfall's typeline, e.g. "Legendary Creature — Human
	// Wizard". Used by deck validation to identify commanders and by
	// the UI to group cards.
	TypeLine string `json:"type_line"`
	// ColorIdentity is the WUBRG letters in this card's color identity.
	// Invariant: values are uppercase single-character strings from
	// the set {"W","U","B","R","G"}. Used to enforce commander color-
	// identity rules.
	ColorIdentity []string `json:"color_identity"`
	// OracleText is Scryfall's up-to-date reminder text. Carried for
	// the client's card-detail panel; the parser also sniffs it for
	// partner / companion clauses so decks using those mechanics can
	// be rejected as unsupported rather than silently accepted.
	OracleText string `json:"oracle_text"`
	// Legalities maps format name → status ("legal", "not_legal",
	// "banned", "restricted"). The only key we use at S05 is
	// "commander"; everything else is carried through as-is for
	// future features.
	Legalities map[string]string `json:"legalities"`
	// ImageURIs is Scryfall's multi-size image map. Keys we care
	// about: "small", "normal", "large", "png", "art_crop",
	// "border_crop". The server picks based on Cache.DefaultSize.
	ImageURIs map[string]string `json:"image_uris"`
	// CardFaces is populated for double-faced / split cards. The
	// front face's image_uris is what we surface for now; handling
	// flip state is a UI concern. The face-level oracle_text lets
	// the parser sniff "Partner" on either half of a partner pair.
	CardFaces []CardFace `json:"card_faces"`
	// Power and Toughness are Scryfall's printed values. Strings,
	// not ints, because creatures like Mortivore use "*" (variable).
	// Empty for non-creatures. Carried for combat-damage resolution
	// (S08); the deck importer parses them to ints when stamping
	// game.Card. Added in S08.
	Power     string `json:"power"`
	Toughness string `json:"toughness"`
	// Loyalty is Scryfall's printed starting loyalty for a
	// planeswalker, as a string for the same reason Power is —
	// some walkers print "X" (Nicol Bolas, the Ravager's back face
	// prints a number; a handful of designs do not). Empty for
	// every non-planeswalker. The deck importer parses it to an int
	// when stamping game.Card.StartingLoyalty, which the engine
	// turns into loyalty counters on battlefield entry (CR 306.5b).
	//
	// Carried here — rather than only in the effect catalog — so
	// that planeswalkers outside the opt-in catalog are playable.
	// See issue #274. Added in the #274 fix.
	Loyalty string `json:"loyalty"`
	// Defense is Scryfall's printed defense for a battle (CR 310.4),
	// a string for the same reason Loyalty is. Empty for every
	// non-battle — and, in practice, empty at the TOP LEVEL for
	// every battle too, because every printed battle is a
	// `transform` card and Scryfall puts the number on the front
	// FACE. See CardFace.Defense; this field exists for the day a
	// single-faced battle is printed.
	//
	// Carried here — rather than only in the effect catalog — for
	// the same reason Loyalty is: without it, a battle outside the
	// opt-in catalog enters with zero defense counters and the
	// CR 704.5p state-based action sweeps it into the graveyard
	// before anyone can attack it. That is #274 one card type over.
	// Added in S27.
	Defense string `json:"defense"`
	// ManaCost is the printed casting cost as Scryfall returns it —
	// e.g. "{1}{R}", "{W/U}", "{X}{B}{B}", "{2}{W}{W}". Empty for
	// lands and for cards without a mana cost (e.g. Pact of Negation
	// alternative-cost printings carry their cost in oracle_text).
	// S15 sub-PR 3 parses this at cast time into a ParsedCost for
	// the optional strict-mode validator. Added in S15 sub-PR 1.
	ManaCost string `json:"mana_cost"`
	// ProducedMana is Scryfall's list of colors a permanent can
	// produce via any mana ability. Entries are uppercase single-
	// character mana letters ("W", "U", "B", "R", "G", "C"). A
	// colorless rock like Sol Ring has ["C"]; Birds of Paradise has
	// ["B", "G", "R", "U", "W"]; non-producers have nil / empty.
	// S15's auto-tapper uses this to narrow the candidate set when
	// solving a cost; basic lands get a synthetic ability derived
	// from their TypeLine instead. Added in S15 sub-PR 1.
	ProducedMana []string `json:"produced_mana"`
	// Keywords is Scryfall's list of the keyword abilities printed
	// on this card — "Flying", "First strike", "Flash", and the
	// hundreds of set-specific mechanics ("Prepared", "Waterbend")
	// alongside them. Capitalisation is Scryfall's; the deck
	// importer lowercases and filters to the keywords the engine
	// actually enforces before stamping game.Card.Keywords.
	//
	// Only keywords the card ITSELF has, never ones it grants:
	// Serra's Blessing ("Creatures you control have vigilance")
	// ships an empty array, and Lightning Greaves lists "Equip"
	// but not the haste and shroud it hands out. That is what
	// makes the field safe to read as printed characteristics.
	//
	// For a multi-faced card the array is the UNION over every
	// face — Aang, Swift Savior // Aang and La, Ocean's Fury lists
	// the front face's flash and flying next to the back face's
	// reach and trample, and the faces carry no keyword arrays of
	// their own to disambiguate. deck.printedKeywords narrows the
	// union against the front face's oracle text.
	//
	// Carried here — rather than only in the effect catalog — so
	// that the ~7,500 cards with a combat keyword and no catalog
	// entry stop having inert keywords. See issues #317 / #319 /
	// #320, and #274 / PR #285 for the same fix applied to printed
	// loyalty.
	Keywords []string `json:"keywords"`
	// Colors is Scryfall's computed color list for the card —
	// uppercase single letters from {"W","U","B","R","G"} — which
	// already accounts for color indicators, Devoid, and hybrid
	// symbols. Empty for colorless. S20 targeting predicates
	// ("non-black creature") read this via game.Card.Colors. Added
	// in S20 sub-PR 1.
	Colors []string `json:"colors"`
}

// CardFace is one printed side of a double-faced / split / flip
// card. Only the fields the server reads are mirrored from Scryfall's
// richer face schema.
type CardFace struct {
	Name       string            `json:"name"`
	TypeLine   string            `json:"type_line"`
	OracleText string            `json:"oracle_text"`
	ImageURIs  map[string]string `json:"image_uris"`
	// Loyalty is the face's printed starting loyalty. Scryfall puts
	// power / toughness / loyalty on the FACE for double-faced
	// cards, leaving the top-level field empty — so without this a
	// transforming planeswalker (Nissa, Vastwood Seer // Sage
	// Animist) would resolve with no loyalty at all. The importer
	// falls back to the first face that prints one.
	Loyalty string `json:"loyalty"`
	// Defense is the FACE's printed defense (CR 310.4). This is
	// where every printed battle's number actually lives: the
	// layout is `transform`, the top-level `defense` is null, and
	// the front face carries "6" / "5" / "4". Empty for every other
	// face. Added in S27.
	Defense string `json:"defense"`
	// ManaCost is the FACE's printed casting cost. This is the field
	// that was missing, and its absence is why every transform and
	// modal DFC imported for free: Scryfall sets the TOP-LEVEL
	// mana_cost to null on those layouts and puts the real cost here
	// (1,203 transform + 300 modal_dfc byName keys), while ParseCost
	// of the empty string succeeds and yields the zero cost. See
	// ADR 0034.
	ManaCost string `json:"mana_cost"`
	// Colors is the face's own colour list. Populated on
	// transform / modal_dfc faces and null on adventure / split
	// ones, which is why the importer resolves colours as
	// colors → color_indicator → derived-from-cost rather than
	// trusting any single field.
	Colors []string `json:"colors"`
	// ColorIndicator is CR 105.2c's coloured dot — the only colour
	// signal on a face that has no mana cost at all, which is every
	// transform back face (Jace, Telepath Unbound prints a blue dot
	// and no cost). Without it those faces read as colourless.
	ColorIndicator []string `json:"color_indicator"`
	// Power and Toughness are the face's printed stats, strings for
	// the same reason the top-level ones are ("*" on Mortivore-class
	// designs). Empty for non-creature faces.
	Power     string `json:"power"`
	Toughness string `json:"toughness"`
}

// Index is a read-only map from card UUID → Card, built at server
// start from the Scryfall bulk dump. Safe for concurrent reads.
//
// Zero-value Index is a valid "empty" index — Get always returns
// (Card{}, false). Useful when the dump file hasn't been downloaded
// yet on a fresh deployment; the server can still run, image
// lookups just 404.
type Index struct {
	mu   sync.RWMutex
	byID map[uuid.UUID]Card
	// byName resolves a case-insensitive card name to the best-match
	// Card for deck imports. Scryfall may ship multiple printings of
	// the same name (different sets); the bulk-dump load path keeps
	// the last one seen, which is usually the most recent printing —
	// good enough for a sandbox. Split / double-faced cards are
	// indexed by both the full printed name ("Fire // Ice") and the
	// front-face name ("Fire") so common deck-list shorthand still
	// resolves.
	byName map[string]Card
	loaded time.Time
	path   string
}

// NewIndex returns an empty Index. Call Load to populate it from a
// Scryfall bulk dump file.
func NewIndex() *Index {
	return &Index{
		byID:   make(map[uuid.UUID]Card),
		byName: make(map[string]Card),
	}
}

// Load replaces the in-memory map with the cards parsed from the
// JSON file at path. Uses a streaming decoder so peak memory stays
// bounded (the dump is a top-level JSON array of ~30k objects).
//
// On success returns the number of cards loaded. If the file does
// not exist, returns os.ErrNotExist unchanged so callers can decide
// whether a missing dump is fatal or merely a "not refreshed yet"
// log warning.
func (i *Index) Load(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)

	// Top-level must be a JSON array — step past the opening bracket.
	tok, err := dec.Token()
	if err != nil {
		return 0, fmt.Errorf("read opening token: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '[' {
		return 0, fmt.Errorf("expected top-level JSON array, got %v", tok)
	}

	loaded := make(map[uuid.UUID]Card, 40_000) // ballpark for default_cards
	byName := make(map[string]Card, 40_000)
	for dec.More() {
		var c Card
		if err := dec.Decode(&c); err != nil {
			return 0, fmt.Errorf("decode card: %w", err)
		}
		if c.ID == uuid.Nil {
			continue // a record without an ID is useless; skip rather than reject
		}
		loaded[c.ID] = c
		// Build the name index. Index both the full printed name and
		// the front-face name for split / double-faced cards, since
		// deck exports vary ("Fire // Ice" vs "Fire").
		//
		// Top-level inserts prefer playable printings over non-playable
		// ones (tokens, art-series, memorabilia). With ties, last-write-
		// wins so reprints of the same playable card carry the most
		// recent legalities. Without this guard, Scryfall's eternalize
		// token records for "Champion of Wits" (5 of them, layout=token,
		// commander=not_legal) shadow the legitimate printings and
		// reject valid decks.
		//
		// Face inserts are guarded: if an existing byName entry's
		// canonical Name already matches the key we're inserting, don't
		// overwrite. This blocks art-series / double-faced-token
		// printings (whose name is "X // X" with two identical faces)
		// from hijacking the legitimate "X" entry via the face loop.
		key := normalizeName(c.Name)
		if existing, ok := byName[key]; !ok || preferIncoming(existing, c) {
			byName[key] = c
		}
		if len(c.CardFaces) > 0 {
			for _, face := range c.CardFaces {
				if face.Name == "" {
					continue
				}
				faceKey := normalizeName(face.Name)
				if existing, ok := byName[faceKey]; ok && normalizeName(existing.Name) == faceKey {
					continue
				}
				if existing, ok := byName[faceKey]; !ok || preferIncoming(existing, c) {
					byName[faceKey] = c
				}
			}
		}
	}
	// Consume the closing bracket; a decoder that swallows it would
	// also swallow trailing junk, so we check.
	if _, err := dec.Token(); err != nil && !errors.Is(err, io.EOF) {
		return 0, fmt.Errorf("read closing token: %w", err)
	}

	i.mu.Lock()
	i.byID = loaded
	i.byName = byName
	i.loaded = time.Now().UTC()
	i.path = path
	i.mu.Unlock()
	return len(loaded), nil
}

// Get returns the card for id, or (zero, false) if no such card is
// loaded.
func (i *Index) Get(id uuid.UUID) (Card, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	c, ok := i.byID[id]
	return c, ok
}

// FindByName resolves a card name to a Card. Matching is case- and
// whitespace-insensitive; for split / double-faced cards both the
// full printed name ("Fire // Ice") and the front-face name ("Fire")
// resolve to the same Card. Returns (zero, false) on miss — the
// deck parser turns that into a structured "unknown card" error.
//
// Fallback: if the exact key misses and the input contains a slash
// separator, retry on the portion before the first slash. This
// handles deck-export variants that use "Stump Stomp / Burnwillow
// Clearing" (single slash) while Scryfall's canonical form is
// "Stump Stomp // Burnwillow Clearing". The front-face name is
// already indexed during Load, so the fallback hits the same Card.
func (i *Index) FindByName(name string) (Card, bool) {
	key := normalizeName(name)
	if key == "" {
		return Card{}, false
	}
	i.mu.RLock()
	defer i.mu.RUnlock()
	if c, ok := i.byName[key]; ok {
		return c, true
	}
	if slash := strings.IndexByte(key, '/'); slash > 0 {
		front := strings.TrimRight(key[:slash], " ")
		if c, ok := i.byName[front]; ok {
			return c, true
		}
	}
	return Card{}, false
}

// Search returns up to limit cards whose name contains q, ranked
// exact match first, then prefix match, then substring — each group
// alphabetical so the ordering is stable between calls. Matching is
// case- and whitespace-insensitive, reusing the same normalisation
// as FindByName.
//
// This is a linear scan of ~30k names. That is fine for its one
// caller (the dev card spawner, one request per keystroke-debounce
// on a single-user preview box) and deliberately not built into an
// index: a prefix trie would be dead weight in production, where
// this code path does not exist.
//
// An empty or whitespace-only q returns nil rather than the whole
// index.
func (i *Index) Search(q string, limit int) []Card {
	key := normalizeName(q)
	if key == "" || limit <= 0 {
		return nil
	}
	i.mu.RLock()
	defer i.mu.RUnlock()

	var exact, prefix, contains []Card
	for name, c := range i.byName {
		switch {
		case name == key:
			exact = append(exact, c)
		case strings.HasPrefix(name, key):
			prefix = append(prefix, c)
		case strings.Contains(name, key):
			contains = append(contains, c)
		}
	}
	byName := func(s []Card) {
		sort.Slice(s, func(a, b int) bool { return s[a].Name < s[b].Name })
	}
	byName(exact)
	byName(prefix)
	byName(contains)

	out := make([]Card, 0, limit)
	for _, group := range [][]Card{exact, prefix, contains} {
		for _, c := range group {
			if len(out) == limit {
				return out
			}
			out = append(out, c)
		}
	}
	return out
}

// preferIncoming reports whether `incoming` should replace `existing`
// under the same byName key. Playable printings (real sets, real
// layouts) always win against non-playable records (tokens, art
// series, memorabilia, minigame, vanguard). Among same-class records
// the incoming wins — preserving the existing "most recent printing
// wins" behavior for legitimate reprints.
func preferIncoming(existing, incoming Card) bool {
	eP := isPlayablePrint(existing)
	iP := isPlayablePrint(incoming)
	if iP && !eP {
		return true
	}
	if !iP && eP {
		return false
	}
	return true
}

// isPlayablePrint reports whether c is a printing a player could
// legitimately put in a deck. Tokens, art series, and minigame /
// vanguard cards are not playable; they should not be the canonical
// byName entry for a given card name.
//
// reversible_card joined the list with ADR 0034. All 81 such
// printings — the Secret Lair double-sided novelty run — carry
// mana_cost: null and type_line: null, and 71 byName keys currently
// resolve to one. The face-insert guard keeps the real single-faced
// entries safe today because those records are all "X // X" doubled
// keys, but a record with no cost and no type belongs with the other
// placeholders rather than one guard away from being a free spell.
func isPlayablePrint(c Card) bool {
	switch c.Layout {
	case "token", "double_faced_token", "art_series", "reversible_card":
		return false
	}
	switch c.SetType {
	case "token", "art_series", "memorabilia", "minigame", "vanguard":
		return false
	}
	return true
}

// normalizeName canonicalises a card name for case-insensitive
// lookup. Trims leading/trailing whitespace, collapses internal
// whitespace runs to a single space, and lowercases. Keeps
// punctuation (apostrophes, commas, hyphens) intact — "Jace, the
// Mind Sculptor" and "Sol Ring" both survive.
func normalizeName(s string) string {
	// Trim edges.
	start, end := 0, len(s)
	for start < end && isASCIISpace(s[start]) {
		start++
	}
	for end > start && isASCIISpace(s[end-1]) {
		end--
	}
	s = s[start:end]
	if s == "" {
		return ""
	}
	// Collapse internal whitespace runs and lowercase ASCII.
	out := make([]byte, 0, len(s))
	prevSpace := false
	for i := 0; i < len(s); i++ {
		b := s[i]
		if isASCIISpace(b) {
			if !prevSpace {
				out = append(out, ' ')
			}
			prevSpace = true
			continue
		}
		prevSpace = false
		if b >= 'A' && b <= 'Z' {
			b += 'a' - 'A'
		}
		out = append(out, b)
	}
	return string(out)
}

func isASCIISpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// Count returns the number of cards currently in the index.
func (i *Index) Count() int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return len(i.byID)
}

// LoadedAt returns the timestamp at which Load last succeeded, or
// the zero time if Load has never been called.
func (i *Index) LoadedAt() time.Time {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.loaded
}

// Put inserts a card into the index, updating both the ID and
// name-keyed maps. Intended for tests that want a small in-memory
// index without a fixture file; production code uses Load to
// populate from the Scryfall bulk dump.
//
// Not safe to call concurrently with Load — the caller is
// responsible for not racing index construction against readers.
func (i *Index) Put(c Card) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.byID[c.ID] = c
	if key := normalizeName(c.Name); key != "" {
		if existing, ok := i.byName[key]; !ok || preferIncoming(existing, c) {
			i.byName[key] = c
		}
	}
	for _, face := range c.CardFaces {
		if face.Name == "" {
			continue
		}
		key := normalizeName(face.Name)
		if key == "" {
			continue
		}
		// See Load: face inserts don't clobber an entry whose
		// canonical Name matches the key. Prevents art-series /
		// same-named-both-faces records from shadowing a legitimate
		// single-card printing.
		if existing, ok := i.byName[key]; ok && normalizeName(existing.Name) == key {
			continue
		}
		if existing, ok := i.byName[key]; !ok || preferIncoming(existing, c) {
			i.byName[key] = c
		}
	}
}

// ImageURI returns the best image URL for c at the given size.
// Falls back to the front face's image if c is a double-faced card,
// and to "normal" if the requested size is missing.
func ImageURI(c Card, size string) string {
	return ImageURIForFace(c, size, 0)
}

// ImageURIForFace is ImageURI for a specific printed face (ADR
// 0034). Face 0 is the front and behaves exactly as ImageURI always
// has, so every existing caller is unaffected.
//
// The order is deliberate and differs per face. For the FRONT face
// the top-level map wins, because that is where Scryfall puts the
// art for single-faced and split-style layouts; the face map is the
// fallback for the transform / modal_dfc shape, whose top-level
// image_uris is absent entirely.
//
// For a BACK face there is no top-level candidate that could
// possibly be right — the top-level image of a transform card is the
// front — so a missing face image returns EMPTY rather than falling
// back and serving the wrong side. A 404 is the honest answer; a
// picker showing the same art twice is not.
func ImageURIForFace(c Card, size string, face int) string {
	pick := func(m map[string]string) string {
		if uri, ok := m[size]; ok && uri != "" {
			return uri
		}
		if uri, ok := m["normal"]; ok && uri != "" {
			return uri
		}
		return ""
	}
	if face <= 0 {
		if uri := pick(c.ImageURIs); uri != "" {
			return uri
		}
		if len(c.CardFaces) > 0 {
			return pick(c.CardFaces[0].ImageURIs)
		}
		return ""
	}
	if face < len(c.CardFaces) {
		return pick(c.CardFaces[face].ImageURIs)
	}
	return ""
}
