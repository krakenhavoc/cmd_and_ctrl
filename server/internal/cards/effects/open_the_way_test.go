package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// open_the_way_test.go — #2581's catalog card: "X can't be greater than
// the number of players in the game. Reveal cards from the top of your
// library until you reveal X land cards. Put those land cards onto the
// battlefield tapped and the rest on the bottom of your library in a
// random order."

const openTheWayOracle = "e07297c7-83bc-4162-bbb1-362bc737efe0"

// stackOpenTheWayLibrary puts, top first: Bear, Forest, Shock, Island,
// Plains on top of the caster's library and returns their IDs in that
// order.
func stackOpenTheWayLibrary(p *game.Player) []uuid.UUID {
	cards := []game.Card{
		{Name: "Bear", TypeLine: "Creature — Bear"},
		{Name: "Forest", TypeLine: "Basic Land — Forest"},
		{Name: "Shock", TypeLine: "Instant"},
		{Name: "Island", TypeLine: "Basic Land — Island"},
		{Name: "Plains", TypeLine: "Basic Land — Plains"},
	}
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushTop(c)
		ids[i] = c.InstanceID
	}
	return ids
}

func castOpenTheWay(t *testing.T, g *game.Game, x int) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Open the Way", TypeLine: "Sorcery", ManaCost: "{X}{G}{G}",
		OracleID: openTheWayOracle, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{XValue: x})
}

func TestOpenTheWayRevealsUntilXLandsAndPutsThemTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := stackOpenTheWayLibrary(me)
	if _, err := castOpenTheWay(t, g, 2); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)

	for _, land := range []uuid.UUID{ids[1], ids[3]} {
		c := findBattlefieldCardByID(g, land)
		if c == nil {
			t.Fatalf("a revealed land is not on the battlefield")
		}
		if !c.Tapped || c.Controller != me.ID {
			t.Errorf("%s entered tapped=%v under %v, want tapped under the caster", c.Name, c.Tapped, c.Controller)
		}
	}
	if findBattlefieldCardByID(g, ids[4]) != nil {
		t.Errorf("the reveal went past the second land")
	}
	// The Plains was never revealed and is still on top; the Bear and
	// the Shock went to the bottom.
	top := me.Library.Cards[len(me.Library.Cards)-1]
	if top.InstanceID != ids[4] {
		t.Errorf("top of library is %s, want the unrevealed Plains", top.Name)
	}
	bottom := map[uuid.UUID]bool{me.Library.Cards[0].InstanceID: true, me.Library.Cards[1].InstanceID: true}
	if !bottom[ids[0]] || !bottom[ids[2]] {
		t.Errorf("the revealed nonland cards are not the bottom two")
	}
}

func TestOpenTheWayRefusesXAboveThePlayersInTheGame(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id, err := castOpenTheWay(t, g, 5)
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("X=5 at four players: err = %v, want ErrInvalidParam", err)
	}
	if !me.Hand.Contains(id) {
		t.Fatalf("the refused cast left the hand")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 4}); err != nil {
		t.Fatalf("X=4 at four players: %v", err)
	}
}

// The 2023-05-12 ruling: "players leaving the game in response to the
// spell won't affect the number of land cards you'll find." And the
// announced X survives a restore.
func TestOpenTheWayKeepsItsXWhenAPlayerLeavesInResponseAndAcrossARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := stackOpenTheWayLibrary(me)
	id, err := castOpenTheWay(t, g, 3)
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	for _, seat := range []int{1, 2} {
		if err := g.Concede(g.Seats[(g.Turn.ActiveSeat+seat)%4].ID); err != nil {
			t.Fatalf("Concede: %v", err)
		}
	}
	if got, _ := g.SpellXCeiling(me.ID, openTheWayOracle); got != 2 {
		t.Fatalf("ceiling after two players left = %d, want 2", got)
	}

	g = restoreRoundTrip(t, g, true)
	if got := g.StackMeta[id].XValue; got != 3 {
		t.Fatalf("XValue after a restore = %d, want the announced 3", got)
	}
	passPriorityAroundTable(t, g)
	for _, land := range []uuid.UUID{ids[1], ids[3], ids[4]} {
		if findBattlefieldCardByID(g, land) == nil {
			t.Errorf("X=3 did not put all three lands onto the battlefield")
		}
	}
}

// A library that runs out first puts every land it revealed.
func TestOpenTheWayWithTooFewLandsPutsWhatItFound(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	ids := stackOpenTheWayLibrary(me)
	if _, err := castOpenTheWay(t, g, 4); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, land := range []uuid.UUID{ids[1], ids[3], ids[4]} {
		if findBattlefieldCardByID(g, land) == nil {
			t.Errorf("a revealed land stayed in the library")
		}
	}
	if len(me.Library.Cards) != 2 {
		t.Errorf("library has %d cards, want the two nonland cards", len(me.Library.Cards))
	}
}
