package game

import (
	"sort"

	"github.com/google/uuid"
)

// proliferate.go is CR 701.34: "Choose any number of permanents
// and/or players with counters on them, then give each another
// counter of each kind already there."
//
// The rule splits cleanly into a CHOICE and an APPLICATION, and only
// the application lives here. applyProliferateLocked takes the chosen
// permanents and players as arguments and does not decide anything:
// the catalog's Proliferate primitive picks them today (a
// deterministic beneficial pick — see the primitive's own comment),
// and an interactive picker, when one lands, feeds the same function
// a player-chosen list without this file changing.
//
// The keyword ACTION — the CR 614 window that "if you would
// proliferate, proliferate twice instead" replaces, and the entry
// point every catalog proliferate goes through — is
// ProliferateForEffect in keyword_action.go (#976). It calls this
// once per time the window settled on.
//
// Two properties the application has to get right:
//
//   - "another counter of EACH KIND already there" — a creature with
//     a +1/+1 and a shield counter gets one of each, and a permanent
//     with no counters at all is not a legal choice and gets nothing.
//     Kinds are snapshotted before any counter is placed, so a
//     replacement that adds a NEW kind mid-loop can't cascade.
//   - The counters go through AddCounterByForEffect, not a raw map
//     write, so the CR 614 replacement pipeline sees them: Doubling
//     Season doubles a proliferated counter exactly as it doubles any
//     other, which is the paper interaction. Since ADR 0056 that is
//     true of the PLAYER half too — poison and energy used to be a bare
//     map write with no window, no event and no layer bump; see
//     player_counters.go.
//   - The PROLIFERATING PLAYER is the placer on every counter it
//     places, on permanents and on players alike (CR 701.34: "GIVE
//     each another counter"). That is what makes Vorinclex halve an
//     opponent's proliferate onto your creature and double your own.

// applyProliferateLocked gives each named permanent and each named
// player one additional counter of every kind they already have —
// ONE proliferate, the application half of CR 701.34 with the choice
// already made and the CR 614 window already settled.
//
// Both lists may be empty — "any number" includes zero, and a
// proliferate with nothing worth choosing is a legal no-op rather
// than an error. IDs that name something that has left the
// battlefield, or a player who is not seated, are skipped: the
// choice is made when the effect starts resolving and the board can
// have moved underneath it.
//
// Caller must hold g.mu (it is an effect-time helper).
func (g *Game) applyProliferateLocked(placer uuid.UUID, cardIDs []uuid.UUID, playerIDs []uuid.UUID) error {
	for _, id := range cardIDs {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		for _, kind := range sortedCounterKinds(c.Counters) {
			if err := g.AddCounterByForEffect(placer, id, kind, 1); err != nil {
				return err
			}
		}
	}
	for _, id := range playerIDs {
		p := g.playerByIDLocked(id)
		if p == nil || p.Eliminated {
			continue
		}
		for _, kind := range sortedCounterKinds(p.Counters) {
			if err := g.AddPlayerCounterByForEffect(placer, id, kind, 1); err != nil {
				return err
			}
		}
	}
	return nil
}

// sortedCounterKinds snapshots the counter names present on a
// counter map, in a stable order. Sorted rather than map order so a
// proliferate over several kinds emits its EventCounterPlaced
// stream identically on every run — map iteration order in Go is
// randomised, and an event log that reorders between runs makes
// replay diffs unreadable.
func sortedCounterKinds(counters map[string]int) []string {
	if len(counters) == 0 {
		return nil
	}
	out := make([]string, 0, len(counters))
	for name, n := range counters {
		if n <= 0 {
			continue
		}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
