package game

import "github.com/google/uuid"

// life_preview.go — the READ-ONLY half of the CR 614 life window, for the
// auto-tapper (#4189).
//
// A mana ability that gains its controller life ("{T}: Add {G}{G}. You gain
// 2 life", The Great Henge) is planned like any other source, but its gain
// goes through the replacement window when it lands, and some windows turn
// a gain into a LOSS: Tainted Remedy's and Plague Drone's "if an opponent
// would gain life, that player loses that much life instead", a halver, a
// "you gain no life" shield. The planner holds itself to one bar for
// everything that costs a player life (painBudgetFor): the click that pays
// for a spell is never the click that takes its player to 0 or below. So it
// asks here what the gain would really become, and books any loss as pain.

// lifeChangePreviewLocked is the life total change a gain of `delta` (a
// positive number) by `source` would really make for `player`, asked
// without touching the game: positive for a gain, zero for none, negative
// for a loss a replacement turned it into.
//
// It walks the SAME gathered effects the real window walks
// (gatherActiveReplacementsLocked) and fires the SAME Replace functions,
// with its own applied set standing in for the per-event map, exactly as
// producedManaPreviewLocked does for a mana production, so the planner and
// the executor cannot drift apart. That is sound for the life window
// because a life replacement rewrites the event and nothing else; an
// effect that would ask its controller a question is skipped, as the real
// window skips it when it cannot pause, and a "your life total can't
// change" or "can't gain life" shield (CR 119.7, CR 614.10) is zero.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) lifeChangePreviewLocked(player, source uuid.UUID, delta int) int {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated || delta == 0 {
		return 0
	}
	ev := &ReplacementEvent{
		Kind:       RepEventLife,
		Source:     source,
		LifePlayer: player,
		LifeDelta:  delta,
	}
	if g.lifeGainForbiddenLocked(ev) {
		return 0
	}
	// ev.ID stays zero, so the gather reads a nil applied map and re-offers
	// everything each pass; this is the set that stands in for it (CR 614.5).
	applied := make(map[ReplacementEffectID]bool, 2)
	for iter := 0; iter < maxReplacementIters; iter++ {
		if ev.Canceled {
			return 0
		}
		chosen := -1
		gathered := g.gatherActiveReplacementsReadOnlyLocked(ev)
		for i := range gathered {
			if applied[gathered[i].id] || asksItsOwnQuestion(gathered[i].effect) {
				continue
			}
			chosen = i
			break
		}
		if chosen < 0 {
			break
		}
		applied[gathered[chosen].id] = true
		if gathered[chosen].effect.Replace != nil {
			// The error is dropped, not logged: emitting is a write, and
			// the real window reports a misbehaving replacement when the
			// life actually changes.
			_ = gathered[chosen].effect.Replace(ev, g, gathered[chosen].source)
		}
	}
	if ev.Canceled {
		return 0
	}
	return ev.LifeDelta
}

// selfLifeGainLossLocked is the life a mana ability's declared "You gain N
// life" rider would COST its controller right now (RiderSelfLifeGain): zero
// on every board without a replacement that turns the gain into a loss,
// and the size of the loss on a board with one. The planner books it as
// pain (tapSource.Pain), so the same budget that keeps a painland from
// taking a player to 0 keeps The Great Henge from doing it under a
// Plague Drone's Rot Fly.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) selfLifeGainLossLocked(controller, source uuid.UUID, a *ManaAbilityShape) int {
	if a == nil || a.RiderSelfLifeGain <= 0 {
		return 0
	}
	if d := g.lifeChangePreviewLocked(controller, source, a.RiderSelfLifeGain); d < 0 {
		return -d
	}
	return 0
}
