package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bladeImported is a card as the deck importer hands it to a game, from a
// Scryfall-shaped record whose oracle id no catalog entry answers for.
func bladeImported(t *testing.T, owner uuid.UUID, name, text string, keywords ...string) game.Card {
	t.Helper()
	oracle := uuid.New()
	if _, ok := Lookup(oracle.String()); ok {
		t.Fatalf("%s must not be in the catalog", name)
	}
	c := deck.ToGameCard(cards.Card{
		ID: uuid.New(), OracleID: oracle, Name: name, Layout: "normal",
		TypeLine: "Creature — Drix", ManaCost: "{3}{U}", Power: "3", Toughness: "3",
		Keywords: keywords, OracleText: text,
	}, false)
	c.InstanceID, c.Owner, c.Controller = uuid.New(), owner, owner
	return c
}

const bladeWarpText = "Flying\nWarp {1}{U} (You may cast this card from your hand for its warp cost. " +
	"Exile this creature at the beginning of the next end step, then you may cast it from exile on a later turn.)"

// The case the old caveat named (#5688): an exiled card with warp that the
// catalog does not run. It can only be in exile by a manual move, and it is
// a legal target all the same; an imported card with no warp is not, and
// neither is a card with warp exiled face down, even one the Blade's
// controller has looked at (CR 406.3a).
func TestBladeOfTheSwarmTucksAnExiledCardWithWarpTheCatalogDoesNotRun(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	warper := bladeImported(t, opp.ID, "Uncatalogued Warper", bladeWarpText, "Flying", "Warp")
	if game.AlternativeCostByKey(game.CatalogKey(warper), "warp") != nil {
		t.Fatal("the fixture must have no catalog warp cost")
	}
	g.WithWriteLock(func() { opp.Hand.PushTop(warper) })
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneExile}, warper.InstanceID); err != nil {
		t.Fatalf("move the card to exile by hand: %v", err)
	}
	plain := bladeImported(t, opp.ID, "Uncatalogued Flier", "Flying", "Flying")
	hidden := bladeImported(t, opp.ID, "Hidden Warper", bladeWarpText, "Flying", "Warp")
	hidden.SetFaceDown(game.FaceDownExiled)
	// The Blade's controller has looked at it, so the target walk judges
	// it and does not merely hide it.
	hidden.AddKnower(me.ID)
	g.WithWriteLock(func() {
		g.Exile.PushTop(plain)
		g.Exile.PushTop(hidden)
	})

	// The stamp is part of a restore point: the restored table answers
	// the same way.
	g = restoreRoundTrip(t, g, true)
	me, opp = g.Seats[0], g.Seats[1]

	avB1ResolveCreature(t, g, me, "Blade of the Swarm", "Creature — Insect Assassin", eoeABBladeOracle)
	passPriorityAroundTable(t, g)
	pick := modePickChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no mode prompt")
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("the second mode has a legal target: %v", err)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the second mode")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(plain.InstanceID)); err == nil {
		t.Fatal("an exiled card without warp is not a legal target")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(hidden.InstanceID)); err == nil {
		t.Fatal("a card exiled face down is not a card with warp")
	}
	pickCard(t, g, me.ID, warper.InstanceID)
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(warper.InstanceID) {
		t.Fatal("the card with warp left exile")
	}
	if bottom, err := opp.Library.Bottom(); err != nil || bottom.InstanceID != warper.InstanceID {
		t.Errorf("the card with warp is on the bottom of its owner's library: %v %v", bottom.Name, err)
	}
	if !g.Exile.Contains(plain.InstanceID) || !g.Exile.Contains(hidden.InstanceID) {
		t.Error("the other exiled cards stay where they are")
	}
}

// Not stronger than printed: with no card with warp in exile the second
// mode has no legal target and can't be chosen, whatever else is there.
func TestBladeOfTheSwarmSecondModeNeedsACardWithWarp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := bladeImported(t, opp.ID, "Uncatalogued Flier", "Flying", "Flying")
	granter := bladeImported(t, opp.ID, "Uncatalogued Granter",
		"Creature cards in your hand have warp {2}{R}. (You may cast a card from your hand for its warp cost.)", "Warp")
	g.WithWriteLock(func() {
		g.Exile.PushTop(plain)
		g.Exile.PushTop(granter)
	})

	blade := avB1ResolveCreature(t, g, me, "Blade of the Swarm", "Creature — Insect Assassin", eoeABBladeOracle)
	passPriorityAroundTable(t, g)
	if pick := modePickChoiceFor(g, me.ID); pick != nil {
		if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err == nil {
			t.Fatal("the second mode has no legal target and can't be chosen")
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, blade, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters %d, want 2", got)
	}
	if !g.Exile.Contains(plain.InstanceID) || !g.Exile.Contains(granter.InstanceID) {
		t.Error("nothing leaves exile")
	}
}
