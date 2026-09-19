package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// horn_of_the_mark_test.go — #952's proof card. Horn of the Mark is
// "look at the top five, you may reveal a creature card from among
// them and put it into your hand, put the rest on the bottom in a
// random order": the whole of the primitive in one sentence.

const hornOfTheMarkOracle = "9b836c32-84f0-41ae-b7d9-67b92c743c60"

// hornTopFive stacks five named cards on top of `p`'s library and
// returns them top-first.
func hornTopFive(p *game.Player, names []struct{ name, typeLine string }) []uuid.UUID {
	ids := make([]uuid.UUID, len(names))
	// PushTop, so the LAST pushed is on top: walk backwards.
	for i := len(names) - 1; i >= 0; i-- {
		id := uuid.New()
		p.Library.PushTop(game.Card{
			InstanceID: id, Name: names[i].name, TypeLine: names[i].typeLine,
			Owner: p.ID, Controller: p.ID,
		})
		ids[i] = id
	}
	return ids
}

var hornFive = []struct{ name, typeLine string }{
	{"Mountain", "Basic Land — Mountain"},
	{"Rohirrim Rider", "Creature — Human Knight"},
	{"Shock", "Instant"},
	{"Eomer of the Riddermark", "Creature — Human Noble"},
	{"Sol Ring", "Artifact"},
}

// hornSwing puts the Horn and `attackers` 2/2s under the active seat,
// attacks `defender` with all of them and locks the declaration in.
func hornSwing(t *testing.T, g *game.Game, me *game.Player, defender uuid.UUID, attackers int) {
	t.Helper()
	pushCatalogPermanent(g, me.ID, "Horn of the Mark", "Legendary Artifact", hornOfTheMarkOracle, false)
	advanceTo(t, g, game.StepDeclareAttackers)
	for i := 0; i < attackers; i++ {
		id := acPermanent(g, me.ID, "Rider", "Creature — Human Knight", 2, 2)
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker %d: %v", i, err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
}

func hornRevealed(g *game.Game, id uuid.UUID) bool {
	for _, ev := range g.Events {
		if ev.Kind == game.EventRevealCards && ev.CardID == id {
			return true
		}
	}
	return false
}

// TestHornOfTheMarkTakesACreatureAndBottomsTheRest is the whole card.
func TestHornOfTheMarkTakesACreatureAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	top := hornTopFive(me, hornFive)
	hornSwing(t, g, me, opp.ID, 2)

	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("two attackers raised no Horn of the Mark prompt")
	}
	// Only the CREATURE cards of the five are offered, and only one
	// may be taken — "you may reveal A creature card".
	want := map[uuid.UUID]bool{top[1]: true, top[3]: true}
	if len(prompt.ChooseCards) != len(want) {
		t.Fatalf("offered %d cards, want the %d creature cards of the five", len(prompt.ChooseCards), len(want))
	}
	for _, id := range prompt.ChooseCards {
		if !want[id] {
			t.Errorf("offered a card that is not a creature card: %s", id)
		}
	}
	if prompt.ChooseMin != 0 || prompt.ChooseMax != 1 {
		t.Errorf("bounds = [%d,%d], want [0,1] for an optional single take", prompt.ChooseMin, prompt.ChooseMax)
	}
	// The look is PRIVATE: nobody else is asked, and nobody else is
	// shown the five.
	if latestChooseCards(g, opp.ID) != nil {
		t.Error("the opponent was asked about the Horn's look")
	}

	handBefore := me.Hand.Size()
	libBefore := me.Library.Size()
	taken := top[3]
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{taken}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	if !me.Hand.Contains(taken) {
		t.Error("the chosen creature card is not in hand")
	}
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want exactly one card taken", handBefore, me.Hand.Size())
	}
	if me.Library.Size() != libBefore-1 {
		t.Errorf("library %d → %d, want exactly one card gone", libBefore, me.Library.Size())
	}
	// The card taken is REVEALED; the four it was chosen from are not.
	if !hornRevealed(g, taken) {
		t.Error("the taken card was not revealed — the card prints \"you may reveal\"")
	}
	for _, id := range top {
		if id != taken && hornRevealed(g, id) {
			t.Errorf("a card that was only LOOKED at was revealed: %s", id)
		}
	}
	// The rest went to the bottom: none of them is in the top five any
	// more, and all four are still in the library.
	// Zone.Cards is bottom-first (PushTop appends), so the bottom four
	// are the first four entries.
	bottom := me.Library.Cards[:4]
	rest := map[uuid.UUID]bool{}
	for _, id := range top {
		if id != taken {
			rest[id] = true
		}
	}
	for _, c := range bottom {
		if !rest[c.InstanceID] {
			t.Errorf("%q is on the bottom and was not one of the looked-at cards", c.Name)
		}
		delete(rest, c.InstanceID)
	}
	if len(rest) != 0 {
		t.Errorf("%d of the looked-at cards did not reach the bottom", len(rest))
	}
}

// Declining takes nothing and still bottoms the five. "You may".
func TestHornOfTheMarkMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	top := hornTopFive(me, hornFive)
	hornSwing(t, g, me, opp.ID, 2)

	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no Horn of the Mark prompt")
	}
	handBefore, libBefore := me.Hand.Size(), me.Library.Size()
	if err := g.ResolveChooseCards(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if me.Hand.Size() != handBefore {
		t.Errorf("hand %d → %d after declining", handBefore, me.Hand.Size())
	}
	if me.Library.Size() != libBefore {
		t.Errorf("library %d → %d after declining — nothing should leave it", libBefore, me.Library.Size())
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range me.Library.Cards[:5] {
		bottom[c.InstanceID] = true
	}
	for _, id := range top {
		if !bottom[id] {
			t.Errorf("a looked-at card did not reach the bottom: %s", id)
		}
	}
}

// "TWO OR MORE creatures". One attacker is not a trigger.
func TestHornOfTheMarkNeedsTwoAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hornTopFive(me, hornFive)
	hornSwing(t, g, me, opp.ID, 1)

	if latestChooseCards(g, me.ID) != nil {
		t.Error("one attacker triggered Horn of the Mark")
	}
}

// "attack A PLAYER": two attackers split across two opponents is not
// two creatures attacking a player, so nothing triggers. The same key
// is what makes three attackers on one player trigger once rather than
// three times.
func TestHornOfTheMarkIsPerDefendingPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	first := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	second := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	hornTopFive(me, hornFive)

	pushCatalogPermanent(g, me.ID, "Horn of the Mark", "Legendary Artifact", hornOfTheMarkOracle, false)
	advanceTo(t, g, game.StepDeclareAttackers)
	a := acPermanent(g, me.ID, "Rider A", "Creature — Human Knight", 2, 2)
	b := acPermanent(g, me.ID, "Rider B", "Creature — Human Knight", 2, 2)
	if err := g.DeclareAttacker(a, first.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(b, second.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)

	if latestChooseCards(g, me.ID) != nil {
		t.Error("one attacker per opponent triggered Horn of the Mark")
	}
}

// Three attackers at one player is ONE trigger, not three.
func TestHornOfTheMarkTriggersOncePerDeclaration(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hornTopFive(me, hornFive)
	hornSwing(t, g, me, opp.ID, 3)

	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID {
			n++
		}
	}
	if n != 1 {
		t.Errorf("three attackers raised %d Horn prompts, want 1", n)
	}
}
