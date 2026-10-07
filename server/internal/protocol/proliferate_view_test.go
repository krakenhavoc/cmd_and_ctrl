package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestProliferatePromptCarriesItsSeatsAndTheSuggestion — #2525. The
// permanents ride Options, the seats on offer and the engine's suggested
// answer ride beside them, and all of it is public: counters are on the
// table, so a seat that is not the chooser sees the same question.
func TestProliferatePromptCarriesItsSeatsAndTheSuggestion(t *testing.T) {
	g := newTwoSeatGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	bear.AddKnowersAll([]uuid.UUID{me.ID, opp.ID})
	g.Battlefield.PushTop(bear)
	if err := g.AddCounter(bear.InstanceID, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayerCounter(opp.ID, game.CounterPoison, 1); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if err := g.ProliferateChoosingForEffect(me.ID, uuid.Nil, nil); err != nil {
			t.Fatalf("ProliferateChoosingForEffect: %v", err)
		}
	})

	for _, viewer := range []*game.Player{me, opp} {
		v := FilterViewFor(ViewOfGame(g), viewer.ID.String())
		if len(v.PendingChoices) != 1 {
			t.Fatalf("%s saw %d choices, want 1", viewer.Name, len(v.PendingChoices))
		}
		c := v.PendingChoices[0]
		if c.Kind != "proliferate" {
			t.Fatalf("kind = %q, want proliferate", c.Kind)
		}
		if c.Chooser != me.ID.String() {
			t.Errorf("chooser = %q, want the proliferating seat", c.Chooser)
		}
		if len(c.Options) != 1 || c.Options[0].InstanceID != bear.InstanceID.String() {
			t.Errorf("%s: options = %+v, want the one countered permanent", viewer.Name, c.Options)
		}
		if len(c.ChoosePlayers) != 1 || c.ChoosePlayers[0] != opp.ID.String() {
			t.Errorf("%s: choose_players = %v, want the seat with poison", viewer.Name, c.ChoosePlayers)
		}
		// Mine is helped by a +1/+1 counter and their poison hurts them.
		want := map[string]bool{bear.InstanceID.String(): true, opp.ID.String(): true}
		if len(c.ChooseSuggested) != 2 || !want[c.ChooseSuggested[0]] || !want[c.ChooseSuggested[1]] {
			t.Errorf("%s: choose_suggested = %v, want %v", viewer.Name, c.ChooseSuggested, want)
		}
		if c.ChooseMax != 2 || c.ChooseMin != 0 {
			t.Errorf("%s: bounds = %d..%d, want 0..2", viewer.Name, c.ChooseMin, c.ChooseMax)
		}
	}
}
