package oracletext

import (
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ability.go maps a catalog row to the printed ability it implements
// (#3830): the stack shows "Whenever another creature you control
// enters, Purphoros deals 2 damage to each opponent." rather than the
// row's engine label "Purphoros, God of the Forge — 2 damage to each
// opponent".
//
// Nothing is written per card. The map is derived from the card's
// Oracle text the first time one of its abilities is asked about, once
// per catalog key, by MatchTriggered and MatchActivated: a triggered
// row's clause, an activated row's whole "cost: effect" line. A row
// with no single clause answers "", and the caller shows the label.
//
// The text comes from a Source: the server's Scryfall index in
// production (cmd/server wires it at boot), the oracle fixtures in
// tests. With no source, every answer is "".

// Card is one card's printed text as a Source answers it. A
// single-faced card has one face.
type Card struct {
	Name  string
	Faces []Face
}

// Face is one printed face.
type Face struct {
	Name string
	Text string
}

// Source answers a base oracle ID with the card's printed text.
type Source func(oracleID uuid.UUID) (Card, bool)

var (
	mu     sync.Mutex
	source Source
	byKey  map[string]*keyAbilities
)

// keyAbilities is one catalog key's rows, by declared label, mapped to
// their printed text. A label two rows share with different answers
// maps to "".
type keyAbilities struct {
	triggered map[string]string
	activated map[string]string
}

// SetSource installs where printed text comes from and forgets every
// map derived from the previous one.
func SetSource(s Source) {
	mu.Lock()
	defer mu.Unlock()
	source, byKey = s, nil
}

// ownRefPrefix is the prefix of game.OwnAbilityRef: a row of the
// object's own definition, not a granted bundle's.
const ownRefPrefix = "own:"

// AbilityText is the printed text of the ability a catalog ref names,
// or "" when it has none of its own: a granted row (its text is the
// grantor's, in quotes), a key with no Oracle text (a token, an
// emblem), or a row no single clause answers.
func AbilityText(ref game.AbilityRef) string {
	if !strings.HasPrefix(ref.Ref, ownRefPrefix) || ref.Name == "" {
		return ""
	}
	ka := abilitiesOf(game.BaseCatalogKey(ref.Key))
	if ka == nil {
		return ""
	}
	switch ref.Slot {
	case game.AbilitySlotTriggered:
		return ka.triggered[ref.Name]
	case game.AbilitySlotActivated:
		return ka.activated[ref.Name]
	}
	return ""
}

// abilitiesOf derives (once) and returns a key's map, nil when there is
// nothing to derive it from.
func abilitiesOf(key string) *keyAbilities {
	mu.Lock()
	defer mu.Unlock()
	if ka, ok := byKey[key]; ok {
		return ka
	}
	ka, final := deriveLocked(key)
	if final {
		if byKey == nil {
			byKey = map[string]*keyAbilities{}
		}
		byKey[key] = ka
	}
	return ka
}

// deriveLocked builds a key's map. final is false when the source had
// no text for the card, so a later ask (once the index has loaded)
// tries again.
func deriveLocked(key string) (ka *keyAbilities, final bool) {
	if source == nil || game.CatalogLookup == nil {
		return nil, false
	}
	base, face := splitFaceKey(key)
	id, err := uuid.Parse(base)
	if err != nil || id.String() != base || face < 0 {
		return nil, true
	}
	def := game.CatalogLookup(key)
	if def == nil || (len(def.Triggered) == 0 && len(def.Activated) == 0) {
		return nil, true
	}
	card, ok := source(id)
	if !ok || len(card.Faces) == 0 {
		return nil, false
	}
	faces := card.Faces
	switch {
	case face > 0 && face < len(faces):
		faces = faces[face : face+1]
	case face > 0:
		return nil, true
	case len(faces) > 1 && game.CatalogLookup(base+"#1") != nil:
		// The back face is registered on its own, so this key is the
		// front face's.
		faces = faces[:1]
	}
	names := []string{card.Name}
	for _, f := range card.Faces {
		names = append(names, f.Name)
	}
	names = CardNames(names...)
	var clauses []Clause
	for _, f := range faces {
		clauses = append(clauses, Clauses(f.Text, names)...)
	}

	trig := make([]string, len(def.Triggered))
	for i, t := range def.Triggered {
		trig[i] = t.Key
	}
	act := make([]string, len(def.Activated))
	for i, a := range def.Activated {
		act[i] = a.Label
	}
	return &keyAbilities{
		triggered: byLabel(trig, MatchTriggered(trig, clauses, names)),
		activated: byLabel(act, MatchActivated(act, clauses, names)),
	}, true
}

// byLabel keys answers by label; a label with two different answers
// has none.
func byLabel(labels, texts []string) map[string]string {
	out := make(map[string]string, len(labels))
	for i, l := range labels {
		if prev, seen := out[l]; seen && prev != texts[i] {
			out[l] = ""
			continue
		}
		out[l] = texts[i]
	}
	return out
}

// splitFaceKey splits "<oracle id>#<face>": face 0 when unsuffixed, -1
// when the suffix is not a face number.
func splitFaceKey(key string) (string, int) {
	base, suffix, ok := strings.Cut(key, "#")
	if !ok {
		return key, 0
	}
	n, err := strconv.Atoi(suffix)
	if err != nil || n < 0 {
		return base, -1
	}
	return base, n
}
