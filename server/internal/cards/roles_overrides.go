package cards

import (
	"cmp"
	"slices"

	"github.com/google/uuid"
)

// roles_overrides.go — the short list of cards whose Tagger roles are
// plainly wrong for our purposes (ADR 0217 §7, override 2). It is
// reviewed in PRs like a catalog card and is expected to stay short: a
// tag that is merely debatable is left alone, because Tagger is edited
// weekly and a debate is not a bug.
//
// An entry names the card by its Scryfall oracle_id (never by name or
// printing) and says why in a comment. Add and Remove are role strings
// (the Role constants). It is applied last, after the Tagger roles and
// the catalog Purpose, so a Remove wins over both.

// roleEdit is one card's override.
type roleEdit struct {
	Add    []string
	Remove []string
}

// roleOverrides is keyed by oracle_id. Empty on purpose until a tag is
// found that is plainly wrong; the first entry arrives with its reason.
var roleOverrides = map[uuid.UUID]roleEdit{}

// ApplyRoleOverrides returns roles with the oracle_id's override list
// applied, in AllRoles order. roles is not modified. A card with no
// override gets its roles back sorted into AllRoles order, so the result
// is the same however the caller built the input.
func ApplyRoleOverrides(oracleID uuid.UUID, roles []string) []string {
	edit, ok := roleOverrides[oracleID]
	out := slices.Clone(roles)
	if ok {
		for _, r := range edit.Add {
			if !slices.Contains(out, r) {
				out = append(out, r)
			}
		}
		out = slices.DeleteFunc(out, func(r string) bool { return slices.Contains(edit.Remove, r) })
	}
	return sortRoles(out)
}

// sortRoles orders roles as AllRoles does and drops repeats. A role not
// in AllRoles sorts last, by name, so it is visible rather than lost.
func sortRoles(roles []string) []string {
	slices.SortFunc(roles, func(a, b string) int {
		ia, ib := slices.Index(AllRoles, a), slices.Index(AllRoles, b)
		switch {
		case ia < 0 && ib < 0:
			return cmp.Compare(a, b)
		case ia < 0:
			return 1
		case ib < 0:
			return -1
		}
		return ia - ib
	})
	return slices.Compact(roles)
}
