package legal

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// teamwork_blight.go — #1703: the bot's payment for an announced
// teamwork or blight cost.
//
// ONE payment each, never every subset, for the reason crewPayment
// gives: the printed number is a floor, and the expansion here is
// already modes × targets × cost payments. A set of optional costs the
// seat cannot pay is not announced (ok == false), so a bot is never
// offered a kicked-by-teamwork cast the engine refuses (#544).

// teamworkBlightPayment is the payment for whichever of the two
// components the announced optional costs demand. Nil lists for a set
// that demands neither, which is every announcement in the catalog
// before #1703.
//
// `mandatory` is the cost the announcement pays in the mandatory slot —
// for an either/or card, the chosen branch, whose "blight 2" (Wild
// Unraveling, ADR 0100 §2) is paid through the same walk.
func (e *enumerator) teamworkBlightPayment(mandatory *game.AdditionalCost, optional []game.AdditionalCost, chosen []int) (team, blight []uuid.UUID, ok bool) {
	teamwork, blightN := 0, 0
	if mandatory != nil {
		teamwork += mandatory.Teamwork
		blightN += mandatory.Blight
	}
	for _, i := range chosen {
		if i < 0 || i >= len(optional) {
			continue
		}
		teamwork += optional[i].Teamwork
		blightN += optional[i].Blight
	}
	if teamwork > 0 {
		// crewPayment IS teamwork's walk: untapped creatures the seat
		// controls, greedy from the biggest effective power down, no
		// summoning-sickness filter (CR 702.194a is crew's sentence).
		// Its candidates are the engine's TeamworkOptionsForEffect.
		team = e.crewPayment(teamwork)
		if team == nil {
			return nil, nil, false
		}
	}
	if blightN > 0 {
		id, found := e.blightPayment(blightN, team)
		if !found {
			return nil, nil, false
		}
		blight = []uuid.UUID{id}
	}
	return team, blight, true
}

// blightPayment picks the creature a blight N goes on: from the
// engine's BlightOptionsForEffect, a creature that SURVIVES N -1/-1
// counters if there is one, and among those the one with the least
// power — the -1/-1 costs it least. With no survivor it still pays
// with the least powerful creature: the cost is legal (CR 701.68b asks
// only that the counters can be put), and a bot that wants the bigger
// spell may spend a token on it. `avoid` keeps a creature already in
// the teamwork set from also being chosen. No printed card has both
// costs, so today it is always empty.
func (e *enumerator) blightPayment(n int, avoid []uuid.UUID) (uuid.UUID, bool) {
	skip := make(map[uuid.UUID]bool, len(avoid))
	for _, id := range avoid {
		skip[id] = true
	}
	type candidate struct {
		id        uuid.UUID
		power     int
		survives  bool
		toughness int
	}
	var pool []candidate
	for _, id := range e.g.BlightOptionsForEffect(e.seat) {
		if skip[id] {
			continue
		}
		c := findBattlefield(e.g, id)
		if c == nil {
			continue
		}
		pool = append(pool, candidate{
			id:        id,
			power:     c.CurrentPower(),
			toughness: c.CurrentToughness(),
			survives:  c.CurrentToughness() > n,
		})
	}
	if len(pool) == 0 {
		return uuid.Nil, false
	}
	sort.SliceStable(pool, func(i, j int) bool {
		if pool[i].survives != pool[j].survives {
			return pool[i].survives
		}
		if pool[i].power != pool[j].power {
			return pool[i].power < pool[j].power
		}
		return pool[i].toughness > pool[j].toughness
	})
	return pool[0].id, true
}
