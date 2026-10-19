package cards

import (
	"sort"
	"strings"

	"github.com/google/uuid"
)

// Printings lists selectable English printings of the same card. Import
// identity is always checked by oracle ID, never by alternate title.
func (i *Index) Printings(id uuid.UUID) []Card {
	i.mu.RLock()
	defer i.mu.RUnlock()
	original, ok := i.byID[id]
	out := []Card{}
	if !ok {
		return out
	}
	for _, c := range i.byID {
		same := c.OracleID == original.OracleID
		if original.OracleID == uuid.Nil {
			same = normalizeName(c.Name) == normalizeName(original.Name)
		}
		if !same || !isPlayablePrint(c) || (c.Lang != "" && c.Lang != "en") || c.SetCode == "" || c.CollectorNo == "" {
			continue
		}
		out = append(out, c)
	}
	sort.Slice(out, func(a, b int) bool {
		x, y := out[a], out[b]
		if x.ReleasedAt != y.ReleasedAt {
			return x.ReleasedAt > y.ReleasedAt
		}
		if x.SetCode != y.SetCode {
			return x.SetCode < y.SetCode
		}
		if x.CollectorNo != y.CollectorNo {
			return x.CollectorNo < y.CollectorNo
		}
		return strings.Compare(x.ID.String(), y.ID.String()) < 0
	})
	return out
}
