package game

import "github.com/google/uuid"

// life_each_keyed.go — #4253 slice 3: "each of these players loses N
// life, then …" with the rest of the batch as data.
//
// LoseLifeEachThenForEffect walks the players one life change at a time,
// each from the previous change's continuation, so a CR 616 ordering
// prompt over one player's loss waits with the others behind it. Its
// step is a closure over the remaining players and the running total.
// LoseLifeEachKeyedForEffect walks the same batch on a data tail
// (lifeTail.each), and finishes with a registered tail: a drain's "you
// gain life equal to the life lost this way" (Gray Merchant of Asphodel,
// Kokusho), so the prompt is a restore point.

// lifeEachRun is the batch, riding the life tail of the player now
// losing life: the players still to lose it, how much each loses, the
// running total and the effect's continuation. Never written after it
// is built, so an undo snapshot may share it.
type lifeEachRun struct {
	Source uuid.UUID          `json:"source"`
	Rest   []uuid.UUID        `json:"rest,omitempty"`
	Amount int                `json:"amount"`
	Lost   int                `json:"lost,omitempty"`
	Then   *keyedTailSnapshot `json:"then,omitempty"`
}

// loseLifeEachKeyedStepLocked is loseLifeEachStepLocked with the rest of
// the batch as data.
//
// Caller must hold g.mu.
func (g *Game) loseLifeEachKeyedStepLocked(r *lifeEachRun) error {
	rest := r.Rest
	for len(rest) > 0 {
		next := rest[0]
		rest = rest[1:]
		if p := g.playerByIDLocked(next); p == nil || p.Eliminated {
			continue
		}
		return g.changePlayerLifeWithTailLocked(r.Source, next, -r.Amount, &lifeTail{each: &lifeEachRun{
			Source: r.Source, Rest: copyUUIDs(rest), Amount: r.Amount, Lost: r.Lost, Then: r.Then,
		}})
	}
	return g.runKeyedTailLocked(restoreKeyedTail(r.Then), nil, r.Lost)
}

// runLifeEachStepLocked counts the loss that just settled and walks on.
// Only a LOSS counts as life lost this way, as in loseLifeEachStepLocked:
// a replacement that turned the loss into a gain lost nobody anything.
//
// Caller must hold g.mu.
func (g *Game) runLifeEachStepLocked(r *lifeEachRun, applied int) error {
	lost := r.Lost
	if applied < 0 {
		lost -= applied
	}
	return g.loseLifeEachKeyedStepLocked(&lifeEachRun{
		Source: r.Source, Rest: r.Rest, Amount: r.Amount, Lost: lost, Then: r.Then,
	})
}

func cloneLifeEachRun(r *lifeEachRun) *lifeEachRun {
	if r == nil {
		return nil
	}
	out := *r
	out.Rest = copyUUIDs(r.Rest)
	if r.Then != nil {
		t := *r.Then
		t.Carry = copyUUIDs(r.Then.Carry)
		out.Then = &t
	}
	return &out
}
