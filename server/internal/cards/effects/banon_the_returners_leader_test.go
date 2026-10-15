package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// banon_the_returners_leader_test.go — #3769: Banon's graveyard cast is
// limited to cards put there THIS TURN from anywhere other than the
// battlefield, once per turn, on your own turn.

const banonOracle = "1b62e88f-d544-4c01-a5a2-1c185a3c9c1e"

func banonTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	b := pushCatalogPermanent(g, me.ID, "Banon, the Returners' Leader", "Legendary Creature — Human Cleric Rebel", banonOracle, false)
	return g, me, b
}

// banonHandCard puts a {0} card in a seat's hand.
func banonHandCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{0}",
		Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID,
		KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	return id
}

func banonMove(t *testing.T, g *game.Game, owner uuid.UUID, from, to game.ZoneKind, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: from, Owner: ownerFor(from, owner)}, game.ZoneRef{Kind: to, Owner: ownerFor(to, owner)}, id); err != nil {
		t.Fatalf("move %s -> %s: %v", from, to, err)
	}
}

// ownerFor names the owner only for the per-player zones.
func ownerFor(z game.ZoneKind, owner uuid.UUID) uuid.UUID {
	switch z {
	case game.ZoneBattlefield, game.ZoneExile, game.ZoneStack:
		return uuid.Nil
	}
	return owner
}

func TestBanonIsRegisteredAndWhole(t *testing.T) {
	spec, ok := Lookup(banonOracle)
	if !ok {
		t.Fatal("Banon is not registered")
	}
	if spec.Completeness != CompletenessFull {
		t.Errorf("completeness = %v, want Full", spec.Completeness)
	}
	if len(spec.CastPermissions) != 1 || !spec.CastPermissions[0].Filter.PutInGraveyardThisTurn {
		t.Errorf("Banon's permission does not carry the arrival filter: %+v", spec.CastPermissions)
	}
}

// A discarded creature card was put into the graveyard from the hand, so
// it qualifies; a creature that died came from the battlefield, so it
// does not; and a milled one qualifies like a discarded one.
func TestBanonOpensDiscardedAndMilledCreaturesButNotOnesThatDied(t *testing.T) {
	g, me, _ := banonTable(t)

	discarded := banonHandCard(me, "Discarded Bear", "Creature — Bear")
	banonMove(t, g, me.ID, game.ZoneHand, game.ZoneGraveyard, discarded)

	dead := pushCatalogPermanent(g, me.ID, "Dead Wolf", "Creature — Wolf", "", false)
	banonMove(t, g, me.ID, game.ZoneBattlefield, game.ZoneGraveyard, dead)

	milled := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: milled, Name: "Milled Elk", TypeLine: "Creature — Elk", ManaCost: "{0}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})

	spell := banonHandCard(me, "Milled Sorcery", "Sorcery")
	banonMove(t, g, me.ID, game.ZoneHand, game.ZoneGraveyard, spell)

	if !enumeratedFor(g, me.ID, discarded) {
		t.Error("a creature card discarded this turn is not offered")
	}
	if !enumeratedFor(g, me.ID, milled) {
		t.Error("a creature card milled this turn is not offered")
	}
	if enumeratedFor(g, me.ID, dead) {
		t.Error("a creature that died this turn (from the battlefield) is offered")
	}
	if enumeratedFor(g, me.ID, spell) {
		t.Error("a noncreature card is offered; Banon casts creature spells")
	}
	if err := muldrothaCast(g, me, dead, ""); err == nil {
		t.Error("a creature that died was cast from the graveyard")
	}
	if err := muldrothaCast(g, me, discarded, ""); err != nil {
		t.Fatalf("the discarded creature: %v", err)
	}
}

// "Once during each of your turns": one cast opens no second.
func TestBanonAllowsOnlyOneCastPerTurn(t *testing.T) {
	g, me, _ := banonTable(t)
	a := banonHandCard(me, "Bear A", "Creature — Bear")
	b := banonHandCard(me, "Bear B", "Creature — Bear")
	banonMove(t, g, me.ID, game.ZoneHand, game.ZoneGraveyard, a)
	banonMove(t, g, me.ID, game.ZoneHand, game.ZoneGraveyard, b)

	if err := muldrothaCast(g, me, a, ""); err != nil {
		t.Fatalf("first cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := muldrothaCast(g, me, b, ""); err == nil {
		t.Error("a second creature spell was cast from the graveyard this turn")
	}
	if enumeratedFor(g, me.ID, b) {
		t.Error("the enumerator still offers a second creature this turn")
	}
}

// "This turn" ends with the turn: a creature put there last turn is no
// longer castable, and a card put there on an opponent's turn is never.
func TestBanonForgetsLastTurnsArrivals(t *testing.T) {
	g, me, _ := banonTable(t)
	old := banonHandCard(me, "Old Bear", "Creature — Bear")
	banonMove(t, g, me.ID, game.ZoneHand, game.ZoneGraveyard, old)
	if !enumeratedFor(g, me.ID, old) {
		t.Fatal("a creature discarded this turn is not offered")
	}

	// Around the table and back to my next main phase.
	advanceToUpkeepOf(t, g, (banonSeat(g, me)+1)%len(g.Seats))
	advanceToUpkeepOf(t, g, banonSeat(g, me))
	advanceToMain(t, g)
	if got, _ := graveyardCard(me, old); got.PutInGraveyardThisTurn {
		t.Error("the arrival record still says 'this turn' a full round later")
	}
	if enumeratedFor(g, me.ID, old) {
		t.Error("last turn's discard is still offered")
	}
	if err := muldrothaCast(g, me, old, ""); err == nil {
		t.Error("a creature put there last turn was cast")
	}
}

func graveyardCard(p *game.Player, id uuid.UUID) (game.Card, bool) {
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

// A creature that died, was exiled and then put into a graveyard from
// exile is a new object whose latest arrival was not from the
// battlefield (CR 400.7), so it qualifies.
func TestBanonJudgesOnlyTheLatestArrival(t *testing.T) {
	g, me, _ := banonTable(t)
	c := pushCatalogPermanent(g, me.ID, "Cat", "Creature — Cat", "", false)
	banonMove(t, g, me.ID, game.ZoneBattlefield, game.ZoneGraveyard, c)
	if enumeratedFor(g, me.ID, c) {
		t.Fatal("a creature that died is offered")
	}
	banonMove(t, g, me.ID, game.ZoneGraveyard, game.ZoneExile, c)
	banonMove(t, g, me.ID, game.ZoneExile, game.ZoneGraveyard, c)
	if !enumeratedFor(g, me.ID, c) {
		t.Error("a card put into the graveyard from exile is not offered")
	}
	if got, _ := graveyardCard(me, c); got.PutInGraveyardFrom != game.ZoneExile {
		t.Errorf("arrival zone = %q, want exile", got.PutInGraveyardFrom)
	}
}

// A countered creature spell goes to the graveyard from the stack.
func TestBanonOpensACounteredCreatureSpell(t *testing.T) {
	g, me, _ := banonTable(t)
	bear := banonHandCard(me, "Countered Bear", "Creature — Bear")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if err := g.CounterSpell(bear, nil); err != nil {
		t.Fatalf("counter: %v", err)
	}
	if !enumeratedFor(g, me.ID, bear) {
		t.Error("a countered creature spell is not offered")
	}
}

// Banon reads a card an opponent's effect put into YOUR graveyard, but
// not a card in an opponent's graveyard.
func TestBanonReadsOnlyYourOwnGraveyard(t *testing.T) {
	g, me, _ := banonTable(t)
	opp := g.Seats[(banonSeat(g, me)+1)%len(g.Seats)]
	theirs := banonHandCard(opp, "Their Bear", "Creature — Bear")
	banonMove(t, g, opp.ID, game.ZoneHand, game.ZoneGraveyard, theirs)
	if enumeratedFor(g, me.ID, theirs) {
		t.Error("a card in an opponent's graveyard is offered")
	}
}

// The permission is closed off your own turn (TimingYourTurnOnly).
func TestBanonDoesNotOpenOnAnOpponentsTurn(t *testing.T) {
	g, me, _ := banonTable(t)
	other := (banonSeat(g, me) + 1) % len(g.Seats)
	advanceToUpkeepOf(t, g, other)
	advanceToMain(t, g)
	late := banonHandCard(me, "Late Bear", "Creature — Bear")
	banonMove(t, g, me.ID, game.ZoneHand, game.ZoneGraveyard, late)
	if enumeratedFor(g, me.ID, late) {
		t.Error("the permission opened on an opponent's turn")
	}
}

// --- the loot ---------------------------------------------------------

func banonAttack(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	opp := g.Seats[(banonSeat(g, me)+1)%len(g.Seats)]
	attacker := pushVanillaCreature(g, me.ID, "Grizzly", 2, 2)
	attackWith(t, g, opp.ID, attacker)
}

func TestBanonLootsOnceWhenYouAttack(t *testing.T) {
	g, me, _ := banonTable(t)
	keep := banonHandCard(me, "Discard Me", "Sorcery")
	handBefore := me.Hand.Size()
	libBefore := me.Library.Size()

	banonAttack(t, g, me)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("attacking did not offer to pay {1}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	answerPayUnless(t, g, me.ID, true)
	answerDiscard(t, g, me.ID, keep)

	if me.Hand.Contains(keep) {
		t.Error("the chosen card was not discarded")
	}
	if !me.Graveyard.Contains(keep) {
		t.Error("the discarded card is not in the graveyard")
	}
	if me.Hand.Size() != handBefore {
		t.Errorf("hand %d, want %d (discard one, draw one)", me.Hand.Size(), handBefore)
	}
	if me.Library.Size() != libBefore-1 {
		t.Errorf("library %d, want %d", me.Library.Size(), libBefore-1)
	}
	if got, _ := graveyardCard(me, keep); got.PutInGraveyardFrom != game.ZoneHand {
		t.Errorf("discard arrived from %q, want hand", got.PutInGraveyardFrom)
	}
}

func TestBanonLootDeclinedChangesNothing(t *testing.T) {
	g, me, _ := banonTable(t)
	banonHandCard(me, "Keep Me", "Sorcery")
	handBefore := me.Hand.Size()
	libBefore := me.Library.Size()

	banonAttack(t, g, me)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("attacking did not offer to pay {1}")
	}
	answerPayUnless(t, g, me.ID, false)
	if me.Hand.Size() != handBefore || me.Library.Size() != libBefore {
		t.Errorf("declining moved cards: hand %d->%d library %d->%d", handBefore, me.Hand.Size(), libBefore, me.Library.Size())
	}
}

// With nothing to discard the cost cannot be paid in full, so nothing is
// offered rather than taking {1} for no card.
func TestBanonOffersNothingWithAnEmptyHand(t *testing.T) {
	g, me, _ := banonTable(t)
	for me.Hand.Size() > 0 {
		c, _ := me.Hand.Top()
		banonMove(t, g, me.ID, game.ZoneHand, game.ZoneLibrary, c.InstanceID)
	}
	banonAttack(t, g, me)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("an empty hand was offered the pay-and-discard loot")
	}
}

// Two attackers are one declaration: one trigger, one offer.
func TestBanonLootsOncePerAttackNotPerAttacker(t *testing.T) {
	g, me, _ := banonTable(t)
	banonHandCard(me, "Fodder", "Sorcery")
	opp := g.Seats[(banonSeat(g, me)+1)%len(g.Seats)]
	a := pushVanillaCreature(g, me.ID, "Bear One", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear Two", 2, 2)
	attackWith(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	prompts := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == me.ID {
			prompts++
		}
	}
	if prompts != 1 {
		t.Errorf("%d pay prompts for one attack, want 1", prompts)
	}
}

func banonSeat(g *game.Game, p *game.Player) int {
	for i, s := range g.Seats {
		if s == p {
			return i
		}
	}
	return -1
}
