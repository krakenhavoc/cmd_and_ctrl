package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// explore_cards_test.go — #2720: the cards that explore (CR 701.44):
// Get Lost's Map tokens and Lodestone Needle // Guidestone Compass.

const getLostOracle = "522dd417-364b-44ab-8ca9-fb55db5f26a6"

// answerExploreQuestion answers the open "put it into your graveyard?"
// prompt.
func answerExploreQuestion(t *testing.T, g *game.Game, chooser uuid.UUID, accept bool) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm {
			if err := g.ResolveConfirm(c.ID, chooser, accept); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			return
		}
	}
	t.Fatal("no explore question is open")
}

func putOnTopForTest(p *game.Player, name, typeLine string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = typeLine
	p.Library.PushTop(c)
	return c.InstanceID
}

// TestGetLostDestroysAndGivesItsControllerTwoMaps — the creature dies
// and its controller, not the caster, gets two Maps.
func TestGetLostDestroysAndGivesItsControllerTwoMaps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	castCatalogSpell(t, g, "Get Lost", "Instant", getLostOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("the target was not destroyed")
	}
	maps := map[uuid.UUID]int{}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Map" {
			maps[c.Controller]++
		}
	}
	if maps[opp.ID] != 2 || maps[me.ID] != 0 {
		t.Errorf("Maps by controller = %v, want two for the destroyed creature's controller", maps)
	}
}

// TestAMapMakesTargetCreatureYouControlExplore — {1}, {T}, sacrifice:
// a nonland on top gives the creature a +1/+1 counter and asks; yes
// bins the card. Only as a sorcery.
func TestAMapMakesTargetCreatureYouControlExplore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	scout := pushVanillaCreature(g, me.ID, "Scout", 1, 1)
	m := pushToken(g, me.ID, MapToken())
	top := putOnTopForTest(me, "Some Sorcery", "Sorcery")

	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, m, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: scout}},
	}); err != nil {
		t.Fatalf("activate the Map: %v", err)
	}
	if g.Battlefield.Contains(m) {
		t.Error("the Map was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, scout); got != 1 {
		t.Errorf("the explorer has %d +1/+1 counters, want 1", got)
	}
	answerExploreQuestion(t, g, me.ID, true)
	if !me.Graveyard.Contains(top) {
		t.Error("the revealed card was not put into the graveyard")
	}
}

// TestAMapIsSorcerySpeed — not while a spell is on the stack.
func TestAMapIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	scout := pushVanillaCreature(g, me.ID, "Scout", 1, 1)
	m := pushToken(g, me.ID, MapToken())
	held := game.NewCard("Test Instant", me.ID)
	held.TypeLine = "Instant"
	me.Hand.PushTop(held)
	if err := g.CastSpell(me.ID, held.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the instant: %v", err)
	}

	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	err := g.ActivateCatalogAbility(me.ID, m, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: scout}},
	})
	if err == nil {
		t.Fatal("a Map activated with a spell on the stack")
	}
	if !g.Battlefield.Contains(m) {
		t.Error("a refused activation sacrificed the Map")
	}
}

// TestLodestoneNeedleTapsAndStunsThenCraftsIntoACompassThatExplores —
// the front's enters trigger taps and stuns; crafted, the back face's
// ability explores, and a land on top goes to hand.
func TestLodestoneNeedleTapsAndStunsThenCraftsIntoACompassThatExplores(t *testing.T) {
	g, me, opp := spendTable(t)
	advanceToMain(t, g)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	castCatalogSpell(t, g, "Lodestone Needle", "Artifact", lodestoneNeedleOracleID, nil)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, theirs); !c.Tapped || c.Counters[game.CounterStun] != 2 {
		t.Errorf("target tapped=%v stun=%d, want tapped with two stun counters", c.Tapped, c.Counters[game.CounterStun])
	}

	needle := pushCraftCard(g, me, transformRow(lodestoneNeedleOracleID, "Lodestone Needle", "Artifact", "{1}{U}",
		"Guidestone Compass", "Artifact", "", "", []string{"U"}))
	rock := pushCatalogPermanent(g, me.ID, "Spare Rock", "Artifact", "", false)
	compass := craftInto(t, g, me, needle, 0, "{C}{C}{U}", "Guidestone Compass", rock)

	scout := pushVanillaCreature(g, me.ID, "Scout", 1, 1)
	land := putOnTopForTest(me, "Forest", "Basic Land — Forest")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, compass.InstanceID, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: scout}},
	}); err != nil {
		t.Fatalf("activate the Compass: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the revealed land did not go to hand")
	}
	if got := plusOneCounters(g, scout); got != 0 {
		t.Errorf("a land explore gave %d counters", got)
	}
}
