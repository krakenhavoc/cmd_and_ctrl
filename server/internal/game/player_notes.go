package game

import (
	"strings"

	"github.com/google/uuid"
)

// player_notes.go — what an object's abilities have recorded for a player
// (#6097).
//
// Notes (notes.go) are words that belong to one object. Some objects keep
// a word PER PLAYER: "As this creature enters, each player chooses war or
// peace. Creatures controlled by players who chose war get +3/+0"
// (Archangel of Strife) is one choice per player that a static reads for
// as long as the permanent stays; "this creature deals X damage to target
// opponent previously dealt damage by it" (Diseased Vermin) is a mark per
// player that a trigger reads as a target restriction. They live on
// Card.PlayerNotes with Card.Notes' lifecycle: per instance, gone when
// the permanent leaves the battlefield (CR 400.7), not copiable
// (CR 707.2), carried by the snapshot.

// PlayerNote is one word recorded for one player by an object.
type PlayerNote struct {
	Player uuid.UUID `json:"player"`
	Note   string    `json:"note"`
}

// copyPlayerNotes is a fresh copy of `in`; nil for an empty list.
func copyPlayerNotes(in []PlayerNote) []PlayerNote {
	if len(in) == 0 {
		return nil
	}
	return append([]PlayerNote(nil), in...)
}

func playerNoteIs(n PlayerNote, player uuid.UUID, note string) bool {
	return n.Player == player && strings.EqualFold(strings.TrimSpace(n.Note), strings.TrimSpace(note))
}

// PlayerNoted reports whether `note` is recorded for `player` on this
// object, compared case-insensitively on trimmed text. A static's Apply
// reads it off its source.
func (c *Card) PlayerNoted(player uuid.UUID, note string) bool {
	if c == nil {
		return false
	}
	for _, n := range c.PlayerNotes {
		if playerNoteIs(n, player, note) {
			return true
		}
	}
	return false
}

// PlayersNoted lists the players `note` is recorded for on this object,
// in the order recorded.
func (c *Card) PlayersNoted(note string) []uuid.UUID {
	if c == nil {
		return nil
	}
	var out []uuid.UUID
	for _, n := range c.PlayerNotes {
		if strings.EqualFold(strings.TrimSpace(n.Note), strings.TrimSpace(note)) {
			out = append(out, n.Player)
		}
	}
	return out
}

// AddPlayerNoteForEffect records `note` for `player` on the battlefield
// permanent `id` ("put a mark": Diseased Vermin's player it dealt damage
// to). A note already recorded for that player, an empty note, a nil
// player and a permanent that has left record nothing. Reports whether
// the note was added. A change to a note a static reads is a new layer
// pass's input, so the layers are marked dirty.
//
// Caller must hold g.mu.
func (g *Game) AddPlayerNoteForEffect(id, player uuid.UUID, note string) bool {
	note = strings.TrimSpace(note)
	if note == "" || player == uuid.Nil {
		return false
	}
	c := findBattlefieldCard(g, id)
	if c == nil || c.PlayerNoted(player, note) {
		return false
	}
	// Always replaced, never appended in place: a clone shares the
	// backing array.
	notes := make([]PlayerNote, 0, len(c.PlayerNotes)+1)
	c.PlayerNotes = append(append(notes, c.PlayerNotes...), PlayerNote{Player: player, Note: note})
	g.layerVersion.Add(1)
	return true
}

// SetPlayerNoteForEffect makes `note` the ONE word recorded for `player`
// on the battlefield permanent `id`, replacing whatever was recorded for
// that player before ("each player chooses war or peace": a player has
// one answer). Reports whether the permanent was found.
//
// Caller must hold g.mu.
func (g *Game) SetPlayerNoteForEffect(id, player uuid.UUID, note string) bool {
	note = strings.TrimSpace(note)
	if note == "" || player == uuid.Nil {
		return false
	}
	c := findBattlefieldCard(g, id)
	if c == nil {
		return false
	}
	notes := make([]PlayerNote, 0, len(c.PlayerNotes)+1)
	for _, n := range c.PlayerNotes {
		if n.Player != player {
			notes = append(notes, n)
		}
	}
	c.PlayerNotes = append(notes, PlayerNote{Player: player, Note: note})
	g.layerVersion.Add(1)
	return true
}

// PlayerNotesForEffect returns a copy of what has been recorded for
// players on the permanent `id`: the live list while it is on the
// battlefield, and the list it had as it last existed there once it has
// left (CR 608.2h). Nil when nothing has been recorded.
//
// Caller must hold g.mu.
func (g *Game) PlayerNotesForEffect(id uuid.UUID) []PlayerNote {
	if c := findBattlefieldCard(g, id); c != nil {
		return copyPlayerNotes(c.PlayerNotes)
	}
	if info, ok := g.LastKnownPermanentForEffect(id); ok {
		return copyPlayerNotes(info.PlayerNotes)
	}
	return nil
}
