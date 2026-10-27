package game

import "github.com/google/uuid"

// exile_face_down_keyed.go — #4253 slice 3: "exile the top N cards of
// your library face down" (#4257, #4261) with the rest of the batch as
// data.
//
// The batch exiles one card at a time, each from the previous leg's
// route continuation, so a leg paused on a CR 616.1 ordering prompt or a
// commander's CR 903.9 choice holds the rest behind it. That continuation
// was a closure over the remaining cards, the landed list and the caller's
// `then` (exileFaceDownLegsLocked, deleted with its last caller in #4253's
// cleanup), so the prompt was not a restore point. The keyed form carries
// the same three things as data on
// the route (zoneRoute.faceDownLegs): the remaining IDs, the landed IDs,
// and the caller's continuation as a registered tail (Necropotence's
// delivery) or a synchronous door's result.

// faceDownExileRun is the rest of a face-down exile batch, riding the
// route of the leg now moving (Leg). Never written after it is built, so
// an undo snapshot may share it.
type faceDownExileRun struct {
	Player uuid.UUID          `json:"player"`
	Leg    uuid.UUID          `json:"leg"`
	Rest   []uuid.UUID        `json:"rest,omitempty"`
	Landed []uuid.UUID        `json:"landed,omitempty"`
	Then   *keyedTailSnapshot `json:"then,omitempty"`
	// report is a synchronous door's result (ExileTopFaceDownForEffect),
	// written when the batch finishes. Never written to a restore point:
	// a batch that paused has returned to the door already.
	report *faceDownExileReport
}

// faceDownExileReport is where a synchronous door reads the batch's end.
type faceDownExileReport struct {
	landed []uuid.UUID
	done   bool
}

func cloneFaceDownExileRun(r *faceDownExileRun) *faceDownExileRun {
	if r == nil {
		return nil
	}
	out := *r
	out.Rest = copyUUIDs(r.Rest)
	out.Landed = copyUUIDs(r.Landed)
	if r.Then != nil {
		t := *r.Then
		t.Carry = copyUUIDs(r.Then.Carry)
		out.Then = &t
	}
	return &out
}

// exileFaceDownLegsKeyedLocked exiles the batch's legs one card at a time,
// with the rest of the batch as data.
//
// Caller must hold g.mu in write mode.
func (g *Game) exileFaceDownLegsKeyedLocked(r *faceDownExileRun) error {
	ids := r.Rest
	for len(ids) > 0 {
		if z := g.findCardZoneLocked(ids[0]); z != nil && z.Kind == ZoneLibrary && z.Owner == r.Player {
			break
		}
		ids = ids[1:]
	}
	if len(ids) == 0 {
		if r.report != nil {
			r.report.landed, r.report.done = r.Landed, true
		}
		return g.runKeyedTailLocked(restoreKeyedTail(r.Then), r.Landed, 0)
	}
	next := ids[0]
	_, err := g.routeCardToZoneLocked(zoneRoute{
		CardID:   next,
		Dst:      ZoneExile,
		Actor:    r.Player,
		FaceDown: FaceDownExiled,
		faceDownLegs: &faceDownExileRun{
			Player: r.Player, Leg: next, Rest: copyUUIDs(ids[1:]), Landed: r.Landed, Then: r.Then, report: r.report,
		},
	})
	return err
}

// runFaceDownExileStepLocked counts the leg that just settled and walks
// on.
//
// Caller must hold g.mu in write mode.
func (g *Game) runFaceDownExileStepLocked(r *faceDownExileRun) error {
	landed := r.Landed
	if g.landedInZoneLocked(r.Leg, ZoneExile, uuid.Nil) {
		landed = append(append(make([]uuid.UUID, 0, len(r.Landed)+1), r.Landed...), r.Leg)
	}
	return g.exileFaceDownLegsKeyedLocked(&faceDownExileRun{
		Player: r.Player, Rest: r.Rest, Landed: landed, Then: r.Then, report: r.report,
	})
}

// ExileTopFaceDownKeyedForEffect is ExileTopFaceDownForEffect for a clause
// that acts on the cards the exile took ("Exile the top card of your library
// face down. Put that card into your hand at the beginning of your next end
// step", #4259), with the continuation as data (#4253 slice 3): `then` is a
// registered tail (RegisterTail), told the cards that landed in exile, top
// card first, as TailOutcome.IDs, once the whole batch is done — at once when
// nothing paused, and from the answer to a CR 616.1 ordering prompt (or a
// CR 903.9 choice) when a leg did. A leg's prompt is then a restore point.
// It also runs, told nothing, when there was nothing to exile (an empty
// library or n of zero); a zero tail is a plain exile. Nothing is exiled
// when the player is not found: the error is returned and the tail does not
// run. (The closure form, ExileTopFaceDownThenForEffect, was deleted in
// #4253's cleanup.)
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileTopFaceDownKeyedForEffect(playerID uuid.UUID, n int, then KeyedTail) error {
	plan, err := g.exileTopFaceDownPlanLocked(playerID, n)
	if err != nil {
		return err
	}
	return g.exileFaceDownLegsKeyedLocked(&faceDownExileRun{Player: playerID, Rest: plan, Then: snapshotKeyedTail(then)})
}

// unknownFaceDownExileKeys is checkEffectKeys' half for a keyed batch.
func unknownFaceDownExileKeys(r *faceDownExileRun) []string {
	if r == nil {
		return nil
	}
	return unknownKeyedTail(r.Then)
}
