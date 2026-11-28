package game

import (
	"strings"

	"github.com/google/uuid"
)

// s82_old_b845_notes.go — a note an ability replaces or takes back (#5636,
// S82 OLD batch 845).
//
// Card.Notes only grows (AddNoteForEffect). Wiitigo's "has blocked or been
// blocked since your last upkeep" is a mark that is set when it blocks and
// taken back at the next upkeep, and Ice Cauldron's "last noted type and
// amount of mana" and "last card exiled with this artifact" are values a new
// activation replaces. Both are one note under a fixed prefix, kept on the
// same Card.Notes list with its lifecycle (per instance, cleared at CR 400.7,
// not copiable, carried by the snapshot): nothing new is stored.

// SetPrefixedNoteForEffect makes `prefix+value` the only note of the
// battlefield permanent `id` that starts with `prefix`. An empty value takes
// the note back. The list is always replaced, never changed in place (clone
// shares the backing array). A permanent that has left notes nothing; reports
// whether the permanent was there.
//
// Caller must hold g.mu.
func (g *Game) SetPrefixedNoteForEffect(id uuid.UUID, prefix, value string) bool {
	c := findBattlefieldCard(g, id)
	if c == nil || prefix == "" {
		return false
	}
	notes := make([]string, 0, len(c.Notes)+1)
	for _, n := range c.Notes {
		if !strings.HasPrefix(n, prefix) {
			notes = append(notes, n)
		}
	}
	if value != "" {
		notes = append(notes, prefix+value)
	}
	if len(notes) == 0 {
		notes = nil
	}
	c.Notes = notes
	return true
}

// PrefixedNoteForEffect reads what SetPrefixedNoteForEffect stored under
// `prefix` for the permanent `id`: its live note while it is on the
// battlefield, the one it had as it last existed there once it has left
// (CR 608.2h). ok is false when nothing is noted under the prefix.
//
// Caller must hold g.mu.
func (g *Game) PrefixedNoteForEffect(id uuid.UUID, prefix string) (value string, ok bool) {
	for _, n := range g.NotesForEffect(id) {
		if v, found := strings.CutPrefix(n, prefix); found {
			return v, true
		}
	}
	return "", false
}
