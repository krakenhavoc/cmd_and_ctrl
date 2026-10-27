package effects

import "github.com/google/uuid"

// carry_ints.go — small numbers on a prompt that carries only IDs (#4253
// sweep). An option pick's Carry is a list of IDs and has no numbers beside
// it, so a continuation that needs a count or an index (the answers an
// earlier prompt of a chain got) writes each as an ID.

// intsToCarry writes each number (0 to 254) as an ID, so they can ride a
// prompt's Carry: n is the ID whose last byte is n+1, which is never
// uuid.Nil, the separator packIDLists uses.
func intsToCarry(ns []int) []uuid.UUID {
	ids := make([]uuid.UUID, len(ns))
	for i, n := range ns {
		ids[i][15] = byte(n + 1)
	}
	return ids
}

// carryToInts is the inverse of intsToCarry.
func carryToInts(ids []uuid.UUID) []int {
	out := make([]int, len(ids))
	for i, id := range ids {
		out[i] = int(id[15]) - 1
	}
	return out
}
