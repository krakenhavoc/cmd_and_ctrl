package game

import (
	"strings"

	"github.com/google/uuid"
)

// s82_old_b846_notes.go — notes an ability takes back (#5636, S82 OLD batch 846).
//
// Card.Notes only grows (AddNoteForEffect). Seraph remembers each creature it
// brought back as one note, "seraph:<id>:<epoch>", and forgets them all once
// it has sacrificed them. A Card.Notes list is replaced, never changed in
// place (clone shares the backing array).

// NotesWithPrefixForEffect is the notes of the permanent `id` that start with
// `prefix`, with the prefix cut off, in the order noted: the live list while it
// is on the battlefield and the one it last had once it has left (CR 608.2h).
//
// Caller must hold g.mu.
func (g *Game) NotesWithPrefixForEffect(id uuid.UUID, prefix string) []string {
	var out []string
	for _, n := range g.NotesForEffect(id) {
		if v, ok := strings.CutPrefix(n, prefix); ok {
			out = append(out, v)
		}
	}
	return out
}

// ClearNotesWithPrefixForEffect takes back every note of the battlefield
// permanent `id` that starts with `prefix`. Reports whether the permanent was
// there.
//
// Caller must hold g.mu.
func (g *Game) ClearNotesWithPrefixForEffect(id uuid.UUID, prefix string) bool {
	c := findBattlefieldCard(g, id)
	if c == nil || prefix == "" {
		return false
	}
	var kept []string
	for _, n := range c.Notes {
		if !strings.HasPrefix(n, prefix) {
			kept = append(kept, n)
		}
	}
	c.Notes = kept
	return true
}
