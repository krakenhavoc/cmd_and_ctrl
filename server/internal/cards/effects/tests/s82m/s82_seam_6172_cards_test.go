package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// s82_seam_6172_cards_test.go — #6172: Wheel of Misfortune's secret
// numbers, 0 or greater with no ceiling, revealed together (CR 608.2d: the
// choices are announced as the effect is applied).

const seam6172WheelOracle = "10aff3f6-8347-4e1b-b8e7-c204c9e46657"

func TestSeam6172WheelIsFull(t *testing.T) {
	spec, ok := effects.Lookup(seam6172WheelOracle)
	if !ok || spec.Name != "Wheel of Misfortune" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Wheel of Misfortune: %+v %v", spec, ok)
	}
}

// seam6172Ballots is the ballots the open round shows `viewer`.
func seam6172Ballots(g *game.Game, viewer string) []protocol.CouncilBallotView {
	var out []protocol.CouncilBallotView
	g.ReadSnapshot(func() {
		for _, c := range protocol.FilterViewFor(protocol.ViewOfGame(g), viewer).PendingChoices {
			if c.CouncilVote != nil {
				out = append(out, c.CouncilVote.Ballots...)
			}
		}
	})
	return out
}

// Four players choose 0, 3, 3 and 1: the two who chose 3 are dealt 3 each,
// and the three who didn't choose 0 discard their hands and draw seven.
func TestWheelOfMisfortuneHighestTakesDamageAndNonLowestWheel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	n := len(g.Seats)
	me := g.Seats[g.Turn.ActiveSeat]
	b, c, d := g.Seats[(g.Turn.ActiveSeat+1)%n], g.Seats[(g.Turn.ActiveSeat+2)%n], g.Seats[(g.Turn.ActiveSeat+3)%n]
	effectstest.AdvanceToMain(t, g)
	oldHands := map[uuid.UUID][]uuid.UUID{}
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			for i := 0; i < 10; i++ {
				p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
			}
			p.Hand.Cards = nil
			for i := 0; i < 2; i++ {
				id := uuid.New()
				p.Hand.PushTop(game.Card{InstanceID: id, Name: "Old Card", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
				oldHands[p.ID] = append(oldHands[p.ID], id)
			}
		}
	})
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.FillPool(me, 2)
	effectstest.CastFromHandForTest(t, g, me, "Wheel of Misfortune", "Sorcery", "{2}{R}", seam6172WheelOracle, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	life := map[uuid.UUID]int{me.ID: me.Life, b.ID: b.Life, c.ID: c.Life, d.ID: d.Life}

	pa := effectstest.AnswerNumber(t, g, me.ID, 0)
	if !pa.NoMax || pa.Min != 0 {
		t.Errorf("the number is 0 or greater with no ceiling: %+v", pa)
	}
	effectstest.AnswerNumber(t, g, b.ID, 3)
	for _, viewer := range []string{c.ID.String(), d.ID.String(), protocol.SpectatorViewerID, ""} {
		if got := seam6172Ballots(g, viewer); len(got) != 0 {
			t.Errorf("viewer %q sees %v before the reveal", viewer, got)
		}
	}
	effectstest.AnswerNumber(t, g, c.ID, 3)
	effectstest.AnswerNumber(t, g, d.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)

	for p, want := range map[*game.Player]int{me: 0, b: 3, c: 3, d: 0} {
		if got := life[p.ID] - p.Life; got != want {
			t.Errorf("%s lost %d life, want %d", p.Name, got, want)
		}
	}
	for _, p := range []*game.Player{b, c, d} {
		if len(p.Hand.Cards) != 7 {
			t.Errorf("%s has %d cards, want 7", p.Name, len(p.Hand.Cards))
		}
		for _, id := range oldHands[p.ID] {
			if !p.Graveyard.Contains(id) {
				t.Errorf("%s did not discard their hand", p.Name)
			}
		}
	}
	for _, id := range oldHands[me.ID] {
		if !me.Hand.Contains(id) {
			t.Error("the player who chose the lowest number discarded")
		}
	}
}

// Everyone choosing the same number is both the highest and the lowest:
// all are dealt it and nobody wheels.
func TestWheelOfMisfortuneAllTiedNobodyWheels(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.FillPool(me, 2)
	effectstest.CastFromHandForTest(t, g, me, "Wheel of Misfortune", "Sorcery", "{2}{R}", seam6172WheelOracle, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	hands, lives := map[uuid.UUID]int{}, map[uuid.UUID]int{}
	for _, p := range g.Seats {
		hands[p.ID], lives[p.ID] = len(p.Hand.Cards), p.Life
	}
	for i := range g.Seats {
		effectstest.AnswerNumber(t, g, g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)].ID, 2)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if lives[p.ID]-p.Life != 2 {
			t.Errorf("%s lost %d life, want 2", p.Name, lives[p.ID]-p.Life)
		}
		if len(p.Hand.Cards) != hands[p.ID] {
			t.Errorf("%s's hand changed %d → %d", p.Name, hands[p.ID], len(p.Hand.Cards))
		}
	}
}
