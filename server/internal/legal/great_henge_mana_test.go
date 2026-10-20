package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// great_henge_mana_test.go is #4185: The Great Henge's "{T}: Add {G}{G}.
// You gain 2 life" is a mana ability (CR 605.1a) whose life gain rides
// after the mana. The auto-tapper used to refuse every source with a
// rider it could not read, so the move list never counted the Henge's
// two mana: a spell only the Henge made affordable was not offered
// until the player tapped the Henge by hand.

const oracleTheGreatHenge = "78427103-9543-41fb-b6d4-72963fe87275"

func theGreatHenge() game.Card {
	return game.Card{
		Name: "The Great Henge", TypeLine: "Legendary Artifact",
		ManaCost: "{7}{G}{G}", OracleID: oracleTheGreatHenge,
	}
}

// TestGreatHengeManaCountsTowardACast: four Forests and The Great Henge
// are six mana, so a six-mana creature is offered, the cast is accepted,
// and paying for it taps the Henge, which gains its 2 life. A seven-mana
// creature is not offered: the Henge counts for two, not more.
func TestGreatHengeManaCountsTowardACast(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	henge := battlefieldCard(g, active, theGreatHenge())
	lands(g, active, "Forest", "Forest", 4)
	six := handCard(active, creature("Green Six", "{5}{G}", 6, 6))
	seven := handCard(active, creature("Green Seven", "{6}{G}", 7, 7))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	if got := castMovesFor(moves, seven); len(got) != 0 {
		t.Errorf("{6}{G} offered off four Forests and The Great Henge: %v", labels(got))
	}
	offered := castMovesFor(moves, six)
	if len(offered) == 0 {
		t.Fatalf("{5}{G} not offered off four Forests and The Great Henge: %v", labels(moves))
	}

	lifeBefore := active.Life
	castAndResolve(t, g, active.ID, offered[0], six)

	tapped, found := false, false
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == henge {
			tapped, found = c.Tapped, true
		}
	}
	if !found {
		t.Fatal("The Great Henge left the battlefield")
	}
	if !tapped {
		t.Error("the cast was paid without tapping The Great Henge")
	}
	if got := active.Life - lifeBefore; got != 2 {
		t.Errorf("life changed by %d paying with The Great Henge, want +2", got)
	}
}
