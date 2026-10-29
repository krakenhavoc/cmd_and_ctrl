package game

import "github.com/google/uuid"

// earned_designations.go is the one sweep behind the player designations
// a static keyword gives "for the rest of the game" once its controller
// controls enough of something:
//
//   - ascend's city's blessing (CR 702.131b): ten or more permanents;
//   - storied's enduring story (CR 702.195a): three or more permanents
//     that are artifacts, Sagas, and/or legendary.
//
// (ADR 0096's 2026-10-08 amendment for the blessing, and its 2026-10-29
// amendment, #4475, which generalised the sweep for the enduring story.)
//
// The two rules are the same sentence with a different count, and the
// same designation rule after it: no rules meaning of its own, any number
// of players may have it at once (CR 702.131c, 702.195b), and the
// continuous effects are reapplied before triggers are checked once a
// player gets it (CR 702.131d, 702.195c). So the engine has one table and
// one walk, and a new "any time you control N …, you get X for the rest of
// the game" keyword is a row here, a bool on Player, and a grant.
//
// The sweep runs from the state-based-action pass (mutations.go). Getting
// a designation is not a state-based action (CR 704 does not list it),
// but that pass is the one place that runs after every action with the
// board settled and before a player receives priority, which is exactly
// when an "any time" static is observable. A permanent that enters as the
// one that makes the count, a token that arrives, a control change, a type
// changed by a layer effect, or the keyword granted to a permanent already
// there all reach it the same way. The caller recomputes the layers after
// a grant, which is CR 702.131d / 702.195c.
//
// Nothing here, or anywhere, takes a designation away.

// earnedDesignation is one row: which keyword earns it, how many
// permanents, which permanents count, and the player's flag. The funcs are
// package-level values, never fields of Game, so the closure census is
// untouched.
type earnedDesignation struct {
	keyword   string
	threshold int
	// counts says whether a permanent its controller controls counts
	// toward the threshold. Each permanent counts once, however many of
	// the qualities it has (an artifact Saga that is legendary is one).
	counts func(c *Card) bool
	// has reads the player's flag.
	has func(p *Player) bool
	// grant is the one write to the flag, which emits the event and
	// reports whether the player got it just now.
	grant func(g *Game, p *Player, source uuid.UUID) bool
}

// earnedDesignations is the table, in a fixed order so the events of two
// designations earned by the same action are emitted in the same order on
// every run (the city's blessing first, as it was before the table).
var earnedDesignations = [...]earnedDesignation{
	{
		keyword:   KeywordAscend,
		threshold: CitysBlessingThreshold,
		counts:    citysBlessingCounts,
		has:       hasCitysBlessing,
		grant:     (*Game).grantCitysBlessingLocked,
	},
	{
		keyword:   KeywordStoried,
		threshold: EnduringStoryThreshold,
		counts:    enduringStoryCounts,
		has:       hasEnduringStory,
		grant:     (*Game).grantEnduringStoryLocked,
	},
}

// earnedDesignationsSweepLocked runs every row. It reports whether anyone
// got anything, so the caller can refresh the layer pass the new
// designation feeds.
//
// Caller must hold g.mu in write mode.
func (g *Game) earnedDesignationsSweepLocked() bool {
	granted := false
	for i := range earnedDesignations {
		if g.earnedDesignationSweepLocked(&earnedDesignations[i]) {
			granted = true
		}
	}
	return granted
}

// earnedDesignationSweepLocked is one row: any player who controls a
// permanent with the row's keyword and at least the threshold of counting
// permanents, and does not have the designation yet, gets it. The source
// named on the event is the first such keyword permanent on the
// battlefield.
//
// One walk over the battlefield, and none at all once every living player
// has the designation, which is where a long game spends its time. The
// per-seat tallies are arrays indexed by seat position, so the common pass
// allocates nothing: this runs after every action in every game, bots
// included.
//
// Caller must hold g.mu in write mode.
func (g *Game) earnedDesignationSweepLocked(d *earnedDesignation) bool {
	if g.Battlefield == nil {
		return false
	}
	var (
		counts  [MaxPlayers]int
		keyed   [MaxPlayers]uuid.UUID
		waiting [MaxPlayers]bool
		someone bool
	)
	for i, p := range g.Seats {
		if i < MaxPlayers && p != nil && !p.Eliminated && !d.has(p) {
			waiting[i], someone = true, true
		}
	}
	if !someone {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		for s, p := range g.Seats {
			if s >= MaxPlayers || !waiting[s] || p.ID != c.Controller {
				continue
			}
			if d.counts(c) {
				counts[s]++
			}
			if keyed[s] == uuid.Nil && HasKeyword(c, d.keyword) {
				keyed[s] = c.InstanceID
			}
			break
		}
	}
	granted := false
	for s, p := range g.Seats {
		if s >= MaxPlayers || !waiting[s] || keyed[s] == uuid.Nil || counts[s] < d.threshold {
			continue
		}
		if d.grant(g, p, keyed[s]) {
			granted = true
		}
	}
	return granted
}
