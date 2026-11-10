package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// riku_of_many_paths_test.go — #4468: Riku's "choose up to X" where X is
// the number of times a mode was chosen for the modal spell that was
// cast (game.RegisterModeCountOfEvent).

const (
	rikuOracle      = "1cfe6e3a-f8c9-4105-ad03-0faf1fcc4626"
	rikuBoltOracle  = "4457ed35-7c10-48c8-9776-456485fdf070"
	rikuAtarkaCmd   = "1943cbb0-2b12-4450-90af-060b8c8627c3"
	rikuFieryConfl  = "3c22e031-4804-4c31-bd3c-c3f29d456b34"
	rikuBirdName    = "Bird"
	rikuTriggerName = "Riku of Many Paths — choose up to X"
)

func rikuPush(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushDiesCreatureForTest(g, owner, "Riku of Many Paths", rikuOracle, "Legendary Creature — Human Wizard", 3, 3)
}

// rikuModePick returns Riku's open mode prompt, passing priority until
// the trigger asks.
func rikuModePick(t *testing.T, g *game.Game, me uuid.UUID) *game.PendingChoice {
	t.Helper()
	if c := modePickChoiceFor(g, me); c != nil {
		return c
	}
	t.Fatal("no mode prompt for Riku's trigger")
	return nil
}

// Two modes chosen for Atarka's Command: up to two of Riku's three.
func TestRikuChoosesUpToTheModesChosenForThatSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	riku := rikuPush(g, me.ID)
	castModal(t, g, "Atarka's Command", "Instant", rikuAtarkaCmd, []int{0, 1}, nil)
	c := rikuModePick(t, g, me.ID)
	if c.ModeMin != 0 || c.ModeMax != 2 {
		t.Errorf("prompt bounds %d..%d, want 0..2 (two modes chosen for that spell)", c.ModeMin, c.ModeMax)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 1, 2}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("three bullets for X = 2: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1, 1}); err == nil {
		t.Fatalf("the same bullet twice was accepted (the fourth ruling, CR 700.2d)")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1, 2}); err != nil {
		t.Fatalf("counter and Bird: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, riku, game.CounterPlusOne); n != 1 {
		t.Errorf("Riku has %d +1/+1 counters, want 1", n)
	}
	if !hasKeywordOnBattlefield(t, g, riku, "trample") {
		t.Errorf("Riku did not gain trample")
	}
	if n, _ := countTokensNamed(g, me.ID, rikuBirdName); n != 1 {
		t.Errorf("%d Bird tokens, want 1", n)
	}
	advanceToNextSeatsTurn(t, g)
	if hasKeywordOnBattlefield(t, g, riku, "trample") {
		t.Errorf("trample outlived the turn")
	}
}

// A mode chosen three times is X = 3 (the third ruling), capped at the
// three bullets; the impulse exile is playable.
func TestRikuCountsARepeatedModeEachTime(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rikuPush(g, me.ID)
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Library Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	castModal(t, g, "Fiery Confluence", "Sorcery", rikuFieryConfl, []int{1, 1, 1}, nil)
	c := rikuModePick(t, g, me.ID)
	if c.ModeMax != 3 {
		t.Fatalf("prompt max %d, want 3: one mode chosen three times", c.ModeMax)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("the exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(top); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the top card is not in exile")
	}
}

// A spell with no modes is not a modal spell: no trigger.
func TestRikuIgnoresANonModalSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rikuPush(g, me.ID)
	start := len(g.Events)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", rikuBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	noModePick(t, g, me.ID, "a non-modal spell")
	passPriorityAroundTable(t, g)
	noModePick(t, g, me.ID, "a non-modal spell, after it resolved")
	if n := countResolved(g, start, rikuTriggerName); n != 0 {
		t.Errorf("Riku's trigger resolved %d times for Lightning Bolt", n)
	}
}

// Only "you cast": an opponent's modal spell does not trigger Riku.
func TestRikuIgnoresAnOpponentsModalSpell(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rikuPush(g, opp.ID)
	castModal(t, g, "Atarka's Command", "Instant", rikuAtarkaCmd, []int{0, 1}, nil)
	noModePick(t, g, opp.ID, "another player's modal spell")
}
