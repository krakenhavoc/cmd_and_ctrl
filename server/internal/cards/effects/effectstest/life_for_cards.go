package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// life_for_cards_test.go — S22's passive draw engines that spend
// life instead of mana: Necropotence, Yawgmoth's Bargain,
// Griselbrand, Vilis and Bloodgift Demon.
//
// Three things these tests are actually about, none of which a
// "hand grew by N" assertion would catch:
//
//   - "Skip your draw step" must skip a STEP and only its
//     controller's. The observable difference is the turn cursor,
//     not the hand size.
//   - Necropotence's exile is FACE DOWN. The assertion that matters
//     is that no seat — including the controller — is a knower of
//     the card while it sits in exile, because exile is a public
//     zone and every other exile in the engine marks the whole
//     table.
//   - The delayed delivery is "at the beginning of YOUR next end
//     step". An opponent's end step in between must not hand the
//     card over early.

const (
	NecropotenceOracle     = "94a844d2-0574-45a7-b347-e0e329767c42"
	YawgmothsBargainOracle = "f7f76f39-a0de-4bda-86b6-0f291892fcec"
	GriselbrandOracle      = "f759d112-76db-4091-a22b-b9f19ab6fa5f"
	VilisOracle            = "7c209753-121b-4859-944b-4e33f885e777"
	BloodgiftDemonOracle   = "e64b184d-9746-4723-a928-d459b5c3ee6c"
)

// --- helpers ------------------------------------------------------

// pushLibraryTopForTest puts a named card on TOP of a player's
// library — the end ExileTopFaceDown and every draw read from.
// pushLibraryCardForTest is the bottom-pushing sibling used by the
// tutor tests, where "first match" wants determinism rather than a
// known top card.
func PushLibraryTopForTest(p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Test",
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}

// soleHandCardForTest empties a player's hand and leaves one named
// card in it, so DiscardRandomForEffect's random pick is the card
// the test wants to watch. The catalog has no "discard this specific
// card" mutator outside an additional cost, and the trigger under
// test does not care which card it was.
func SoleHandCardForTest(p *game.Player, name string) uuid.UUID {
	p.Hand.Cards = nil
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Instant",
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}

// findInZone returns the card with the given instance ID from the
// zone, and whether it was there.
func FindInZone(z *game.Zone, cardID uuid.UUID) (game.Card, bool) {
	if z == nil {
		return game.Card{}, false
	}
	for _, c := range z.Cards {
		if c.InstanceID == cardID {
			return c, true
		}
	}
	return game.Card{}, false
}

// walkSteps advances the turn cursor up to `limit` observable steps,
// recording every (seat, step) pair it lands on, and stops early
// once `done` is satisfied. Returns the trace.
//
// A trace rather than a spot check because "the draw step was
// skipped" is a statement about a step that never happened: there is
// no moment at which a test can look at the cursor and see it.
func WalkSteps(t *testing.T, g *game.Game, limit int, done func() bool) []stepFrame {
	t.Helper()
	trace := []stepFrame{{Seat: g.Turn.ActiveSeat, Step: g.Turn.Step}}
	for i := 0; i < limit; i++ {
		if done() {
			return trace
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep iter %d: %v", i, err)
		}
		trace = append(trace, stepFrame{Seat: g.Turn.ActiveSeat, Step: g.Turn.Step})
	}
	t.Fatalf("turn cursor never reached the wanted position in %d steps", limit)
	return nil
}

// stepFrame is one position of the turn cursor.
type stepFrame struct {
	Seat int
	Step game.Step
}

// sawStep reports whether the trace ever landed on a given seat's
// given step.
func SawStep(trace []stepFrame, seat int, step game.Step) bool {
	for _, f := range trace {
		if f.Seat == seat && f.Step == step {
			return true
		}
	}
	return false
}
