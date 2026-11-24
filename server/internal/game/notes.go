package game

import (
	"strings"

	"github.com/google/uuid"
)

// notes.go — what an object's abilities have noted for it (#6097).
//
// "Note a creature type that hasn't been noted for this Saga" (Long List
// of the Ents), "note one of its creature types that hasn't been noted
// for Volo's Journal", "goad all creatures with a name chosen for this
// enchantment" (Day of the Moon): a list that belongs to one object and
// grows across its abilities' resolutions, read back by a later ability
// of the same object. It lives on Card.Notes, per instance, so a
// permanent that leaves the battlefield and returns is a new object with
// nothing noted (CR 400.7), and a copy notes its own (CR 707.2).

// AddNoteForEffect notes `note` for the battlefield permanent `id`. A
// word already noted (compared case-insensitively on trimmed text) and
// an empty one are not added again; a permanent that has left notes
// nothing. Reports whether the note was added.
//
// Caller must hold g.mu.
func (g *Game) AddNoteForEffect(id uuid.UUID, note string) bool {
	note = strings.TrimSpace(note)
	if note == "" {
		return false
	}
	c := findBattlefieldCard(g, id)
	if c == nil || NotedIn(c.Notes, note) {
		return false
	}
	// Always replaced, never appended in place: a clone shares the
	// backing array.
	notes := make([]string, 0, len(c.Notes)+1)
	c.Notes = append(append(notes, c.Notes...), note)
	return true
}

// NotesForEffect returns a copy of what has been noted for the permanent
// `id`, in the order noted: the live list while it is on the battlefield,
// and the list it had as it last existed there once it has left this turn
// (CR 608.2h). Nil when nothing has been noted.
//
// Caller must hold g.mu.
func (g *Game) NotesForEffect(id uuid.UUID) []string {
	if c := findBattlefieldCard(g, id); c != nil {
		return append([]string(nil), c.Notes...)
	}
	if info, ok := g.LastKnownPermanentForEffect(id); ok {
		return append([]string(nil), info.Notes...)
	}
	return nil
}

// NotedIn reports whether `word` is among `notes`, compared
// case-insensitively on trimmed text.
func NotedIn(notes []string, word string) bool {
	word = strings.TrimSpace(word)
	for _, n := range notes {
		if strings.EqualFold(strings.TrimSpace(n), word) {
			return true
		}
	}
	return false
}
