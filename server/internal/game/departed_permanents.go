package game

import "github.com/google/uuid"

// departed_permanents.go — #6123: what a player controlled when they left
// the game.
//
// Miss Highwater: "For as long as they have a contract counter, when they
// lose the game, for each artifact and creature they controlled, create a
// token that's a copy of it." The trigger fires on the player's loss
// (EventPlayerEliminated, emitted before CR 800.4a removes their objects)
// and resolves afterwards, when the objects are gone. Leaving the game is
// not a zone change (leave_game.go), so no CR 603.10 / 608.2h record of
// the battlefield object survives it, and the token copies have nothing to
// copy. This record is what they controlled as they left, read the way a
// copy and a last-known-information reader would have read it then.

// DepartedPermanent is one permanent a departing player controlled.
type DepartedPermanent struct {
	// Source is the permanent as a token copy copies it (CR 707.2): its
	// copiable values, as other copy effects had changed them
	// (TokenCopySourceForEffect).
	Source Card
	// Types are the card types it had on the battlefield (CR 608.2h): an
	// animated land counts as a creature, an artifact that lost its type
	// does not.
	Types []string
}

// HasType reports whether the permanent had card type t (case-folded) as
// it left.
func (d DepartedPermanent) HasType(t string) bool {
	for _, have := range d.Types {
		if equalFoldASCII(have, t) {
			return true
		}
	}
	return false
}

// recordDepartedPermanentsLocked writes p.DepartedPermanents from the
// battlefield as it stands: every permanent p controls, in battlefield
// order. The layers are read as they stand, as recordFinalBoardLocked
// reads them. Written once; a later call for a player who already has a
// record leaves it.
//
// Caller must hold g.mu.
func (g *Game) recordDepartedPermanentsLocked(p *Player) {
	if p == nil || p.DepartedPermanents != nil || g.Battlefield == nil {
		return
	}
	var ids []uuid.UUID
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Controller == p.ID {
			ids = append(ids, g.Battlefield.Cards[i].InstanceID)
		}
	}
	out := make([]DepartedPermanent, 0, len(ids))
	for _, id := range ids {
		c := findBattlefieldCard(g, id)
		if c == nil {
			continue
		}
		info := permanentInfoOf(c)
		src, ok := g.TokenCopySourceForEffect(id)
		if !ok {
			continue
		}
		out = append(out, DepartedPermanent{Source: src, Types: append([]string(nil), info.Characteristic.Types...)})
	}
	p.DepartedPermanents = out
}

// DepartedPermanentsForEffect is what `player` controlled on the
// battlefield when they left the game, or nil for a player still in it
// (or one who controlled nothing). A copy: the caller may keep it.
// Caller must hold g.mu.
func (g *Game) DepartedPermanentsForEffect(player uuid.UUID) []DepartedPermanent {
	p := g.playerByIDLocked(player)
	if p == nil {
		return nil
	}
	return cloneDepartedPermanents(p.DepartedPermanents)
}

// cloneDepartedPermanents copies the record for Clone and the snapshot.
func cloneDepartedPermanents(src []DepartedPermanent) []DepartedPermanent {
	if src == nil {
		return nil
	}
	out := make([]DepartedPermanent, len(src))
	for i, d := range src {
		out[i] = DepartedPermanent{Source: cloneCard(d.Source), Types: append([]string(nil), d.Types...)}
	}
	return out
}

// departedPermanentSnapshot is a DepartedPermanent on disk: the copy
// source through the card snapshot every zone uses.
type departedPermanentSnapshot struct {
	Source cardSnapshot `json:"source"`
	Types  []string     `json:"types,omitempty"`
}

func snapshotDepartedPermanents(src []DepartedPermanent, cen *ContinuationCensus) []departedPermanentSnapshot {
	if src == nil {
		return nil
	}
	out := make([]departedPermanentSnapshot, len(src))
	for i, d := range src {
		out[i] = departedPermanentSnapshot{Source: snapshotCard(d.Source, cen), Types: append([]string(nil), d.Types...)}
	}
	return out
}

func restoreDepartedPermanents(src []departedPermanentSnapshot) []DepartedPermanent {
	if src == nil {
		return nil
	}
	out := make([]DepartedPermanent, len(src))
	for i := range src {
		out[i] = DepartedPermanent{Source: restoreCard(&src[i].Source), Types: append([]string(nil), src[i].Types...)}
	}
	return out
}
