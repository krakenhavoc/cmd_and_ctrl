package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// call_forth_the_tempest_test.go — #2743: the cast tally's running mana
// value, and the card that reads it after its own two cascades.

const callForthTheTempestOracle = "0623c51a-e829-413f-a3a0-817c0902821e"

// pushBodyForTest seeds a vanilla creature with the given toughness.
func pushBodyForTest(g *game.Game, owner uuid.UUID, name string, toughness int) uuid.UUID {
	id := pushCreatureToBattlefieldForTest(g, owner, name)
	c := &g.Battlefield.Cards[len(g.Battlefield.Cards)-1]
	c.Power, c.Toughness = toughness, toughness
	return id
}

// castWithXForTest pushes a spell into the active seat's hand and
// casts it, with x as its announced X.
func castWithXForTest(t *testing.T, g *game.Game, name, typeLine, manaCost string, x int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Owner: active.ID, Controller: active.ID,
	})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// TestCastTallyAddsEachSpellsManaValueWithItsX — the tally reads the
// spell on the stack as it becomes cast: an X spell counts its X
// (CR 202.3e).
func TestCastTallyAddsEachSpellsManaValueWithItsX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)

	castWithXForTest(t, g, "Test Blaze", "Instant", "{X}{R}", 3)
	if got := g.CastTallyFor(me.ID).ManaValue; got != 4 {
		t.Fatalf("after an X=3 {X}{R} spell the tally is %d, want 4", got)
	}
	castWithXForTest(t, g, "Test Divination", "Instant", "{2}{U}", 0)
	if got := g.CastTallyFor(me.ID).ManaValue; got != 7 {
		t.Errorf("after a second spell of mana value 3 the tally is %d, want 7", got)
	}
	if got := g.CastTallyFor(g.Seats[1].ID).ManaValue; got != 0 {
		t.Errorf("a seat that cast nothing has %d", got)
	}
}

// TestCallForthTheTempestCountsItsCascadedSpells is the deck's reason
// for the card: a spell cast earlier in the turn and both cascade hits
// all count, the Tempest itself does not, and only opponents'
// creatures are dealt the damage.
func TestCallForthTheTempestCountsItsCascadedSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	advanceToMain(t, g)

	// Two cascade hits of mana value 3 each.
	me.Library.Cards = nil
	for i := 0; i < 2; i++ {
		c := game.NewCard("Test Hit", me.ID)
		c.TypeLine = "Instant"
		c.ManaCost = "{2}{U}"
		me.Library.PushTop(c)
	}
	// One spell before the Tempest, mana value 2.
	castWithXForTest(t, g, "Test Opener", "Sorcery", "{1}{R}", 0)
	passPriorityAroundTable(t, g)

	dies := pushBodyForTest(g, opp.ID, "Eight Toughness", 8)
	lives := pushBodyForTest(g, opp.ID, "Nine Toughness", 9)
	mine := pushBodyForTest(g, me.ID, "My Creature", 1)

	castWithCost(t, g, "Call Forth the Tempest", "Sorcery", "{5}{R}{R}{R}", callForthTheTempestOracle)
	hits := 0
	for i := 0; i < 16 && !stackFullyEmpty(g); i++ {
		answerAllMayCast(t, g, me.ID, true)
		for _, c := range g.Exile.Cards {
			if permissionLive(g, g.CastPermissionOnCardByIDForEffect(c.InstanceID), me.ID) {
				if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{FromZone: "exile"}); err != nil {
					t.Fatalf("casting the cascade hit: %v", err)
				}
				hits++
				break
			}
		}
		passPriorityAroundTable(t, g)
	}
	if hits != 2 {
		t.Fatalf("cast %d cascade hits, want 2", hits)
	}
	if got := g.CastTallyFor(me.ID).ManaValue; got != 2+8+3+3 {
		t.Errorf("tally %d, want 16", got)
	}

	// Other spells: 2 + 3 + 3 = 8.
	if battlefieldHas(g, dies) {
		t.Error("an 8-toughness opposing creature survived 8 damage")
	}
	if !battlefieldHas(g, lives) {
		t.Error("a 9-toughness opposing creature died: the Tempest counted itself")
	}
	if !battlefieldHas(g, mine) {
		t.Error("the Tempest damaged its controller's own creature")
	}
}
