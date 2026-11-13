package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abundant Harvest (Strixhaven Mystical Archive, #5503).

const abundantHarvestOracle = "34841228-31ac-48ac-b6d6-486b5e1e07dd"

func TestAbundantHarvestIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(abundantHarvestOracle)
	if !ok {
		t.Fatal("Abundant Harvest is not registered")
	}
	if spec.Name != "Abundant Harvest" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// abundantHarvestCast casts Abundant Harvest as the active seat, whose
// library is replaced by `cards` (the first on top), and answers the
// land-or-nonland question with `pick` (0 land, 1 nonland). It returns
// the caster, the library's IDs in the same order, and the hand size
// before the cast.
func abundantHarvestCast(t *testing.T, pick int, cards ...game.Card) (*game.Game, *game.Player, []uuid.UUID, int) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := trekB40Library(me, cards...)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Abundant Harvest", "Sorcery", abundantHarvestOracle, nil)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, pick)
	passPriorityAroundTable(t, g)
	return g, me, ids, hand
}

// abundantHarvestBottom is the IDs of the bottom n cards of p's library.
func abundantHarvestBottom(p *game.Player, n int) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	for i := 0; i < n && i < len(p.Library.Cards); i++ {
		out[p.Library.Cards[i].InstanceID] = true
	}
	return out
}

func TestAbundantHarvestLandTakesTheFirstLandAndBottomsTheRest(t *testing.T) {
	_, me, ids, hand := abundantHarvestCast(t, 0,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
	)
	if !me.Hand.Contains(ids[2]) || me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d -> %d, Forest in hand %v; want only the first land taken", hand, me.Hand.Size(), me.Hand.Contains(ids[2]))
	}
	bottom := abundantHarvestBottom(me, 2)
	if !bottom[ids[0]] || !bottom[ids[1]] {
		t.Error("the revealed nonland cards are not on the bottom of the library")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != ids[3] {
		t.Error("the unrevealed card is not on top")
	}
}

func TestAbundantHarvestNonlandTakesTheFirstNonland(t *testing.T) {
	_, me, ids, hand := abundantHarvestCast(t, 1,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	if !me.Hand.Contains(ids[2]) || me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d -> %d, Bolt in hand %v; want only the first nonland taken", hand, me.Hand.Size(), me.Hand.Contains(ids[2]))
	}
	bottom := abundantHarvestBottom(me, 2)
	if !bottom[ids[0]] || !bottom[ids[1]] {
		t.Error("the revealed lands are not on the bottom of the library")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != ids[3] {
		t.Error("the unrevealed card is not on top")
	}
}

// A library with no card of the chosen kind is revealed whole, all of it
// stays in the library, and nothing reaches the hand.
func TestAbundantHarvestWithNoCardOfTheChosenKindTakesNothing(t *testing.T) {
	_, me, ids, hand := abundantHarvestCast(t, 0,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want nothing taken", hand, me.Hand.Size())
	}
	for _, id := range ids {
		if !me.Library.Contains(id) {
			t.Error("a revealed card left the library")
		}
	}
	if me.Eliminated {
		t.Error("revealing the whole library lost the game")
	}
}
